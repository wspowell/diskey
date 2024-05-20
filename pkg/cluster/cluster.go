package cluster

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"

	"diskey/pkg/batcher"
	"diskey/pkg/cache"
	"diskey/pkg/discovery"
	"diskey/pkg/rpc"
)

type Option func(clusterClient *Cluster)

func OptionDiscovery(disco discovery.Discovery) func(clusterClient *Cluster) {
	return func(clusterClient *Cluster) {
		clusterClient.disco = disco
	}
}

func OptionLocalhostDiscovery(ports []string) func(clusterClient *Cluster) {
	return func(clusterClient *Cluster) {
		clusterClient.disco = discovery.NewLocalhost(ports)
	}
}

func OptionMemberListPort(portString string) func(clusterClient *Cluster) {
	port, err := strconv.Atoi(portString)
	if err != nil {
		panic(err)
	}
	return func(clusterClient *Cluster) {
		clusterClient.memberListPort = port
	}
}

type clusterMetadata struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type Cluster struct {
	clients        []*rpc.Client
	addresses      []Address
	batchChannel   chan<- *keyRequest
	clientsMutex   sync.RWMutex
	disco          discovery.Discovery
	clusterServer  rpc.Server
	memberList     MemberList
	keyStore       cache.Cache
	memberListPort int
}

func NewCluster(ctx context.Context, host string, port string, options ...Option) *Cluster {
	cacheConfig := cache.Config{}
	keyStore, err := cache.New(context.WithoutCancel(ctx), cacheConfig)
	if err != nil {
		log.Ctx(ctx).Err(err).Send()
		panic(err)
	}

	const batchSize = 1000 // TODO: What is an optimal setting? Expose as config?

	cluster := &Cluster{
		clientsMutex: sync.RWMutex{},
		clients:      []*rpc.Client{},
		addresses: []Address{
			{
				Host: host,
				Port: port,
				Slot: Slot(host + ":" + port),
			},
		},
		disco:          discovery.NewLocalhost([]string{}),
		memberListPort: 7949,
		keyStore:       keyStore,
	}

	for index := range options {
		options[index](cluster)
	}

	batchChannel := batcher.Run(batchSize, func(batch []*keyRequest) {
		cluster.runBatch(ctx, batch)
	})
	cluster.batchChannel = batchChannel

	clusterServer := rpc.NewServer(host, port)
	listener, listenErr := clusterServer.Listen(ctx)
	if listenErr.IsErr() {
		log.Ctx(ctx).Err(listenErr).Send()
		panic(listenErr)
	}

	handlerErr := clusterServer.RegisterHandler(ctx, ClusterCommandRpcHandlers{
		Cluster: cluster,
		// getKeyHandler: cluster.getHandler,
	})
	if handlerErr.IsErr() {
		panic("failed registering cluster commands: " + handlerErr.Error())
	}

	go clusterServer.AcceptConnections(ctx, listener)

	cluster.clusterServer = clusterServer

	metadata, err := json.Marshal(clusterMetadata{
		Host: host,
		Port: port,
	})
	if err != nil {
		panic("failed marshalling metadata: " + err.Error())
	}

	cluster.memberList = NewMemberList(ctx, cluster.disco, metadata,
		MemberListOptionHost(host),
		MemberListOptionPort(cluster.memberListPort),
		MemberListOptionEventCallbacks(ctx, cluster.onJoin, cluster.onLeave, func(ctx context.Context, node *memberlist.Node) {}),
	)

	return cluster
}

func (self *Cluster) Close() error {
	return self.memberList.Shutdown(time.Second)
}

func (self *Cluster) NumClients() int {
	self.clientsMutex.RLock()
	numClients := len(self.clients)
	self.clientsMutex.RUnlock()
	return numClients
}

func (self *Cluster) onJoin(ctx context.Context, node *memberlist.Node) {
	var metadata clusterMetadata
	if err := json.Unmarshal(node.Meta, &metadata); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to unmarshal cluster metadata")
		return
	}

	host := metadata.Host
	port := metadata.Port

	if host == self.clusterServer.Host() && port == self.clusterServer.Port() {
		// This is me.
		log.Ctx(ctx).Debug().Str("host", host).Str("port", port).Msg("this is me")
		return
	}

	self.clientsMutex.Lock()
	defer self.clientsMutex.Unlock()
	for clientsIndex := range self.clients {
		if self.clients[clientsIndex].Host() == host &&
			self.clients[clientsIndex].Port() == port {
			log.Ctx(ctx).Debug().Str("host", host).Str("port", port).Msg("already connected")
			return
		}
	}

	newClient := rpc.NewClient(host, port)
	if connectErr := newClient.Connect(ctx); connectErr.IsErr() {
		// TODO: This needs retry logic to eventually self-heal.
		log.Ctx(ctx).Err(connectErr).Msg("failed to connect new client")
		return
	}
	log.Ctx(ctx).Info().Str("self", self.clusterServer.Address()).Str("host", host).Str("port", port).Msg("client connected")
	self.clients = append(self.clients, newClient)
	self.addresses = append(self.addresses, Address{
		Host: newClient.Host(),
		Port: newClient.Port(),
		Slot: Slot(newClient.Host() + ":" + newClient.Port()),
	})
}

func (self *Cluster) onLeave(ctx context.Context, node *memberlist.Node) {
	self.clientsMutex.Lock()
	defer self.clientsMutex.Unlock()

	host := node.Addr.String()
	port := strconv.Itoa(int(node.Port))

	index := slices.IndexFunc(self.clients, func(client *rpc.Client) bool {
		return client.Host() == host && client.Port() == port
	})

	if index >= 0 {
		self.clients[index].Disconnect(ctx)
		self.clients = slices.Delete(self.clients, index, index)
	}
}

func (self *Cluster) getClientByHostPort(host string, port string) *rpc.Client {
	self.clientsMutex.RLock()
	defer self.clientsMutex.RUnlock()
	for index := range self.clients {
		if self.clients[index].Host() == host && self.clients[index].Port() == port {
			return self.clients[index]
		}
	}
	return nil
}
