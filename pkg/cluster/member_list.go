package cluster

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"

	"diskey/pkg/discovery"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/memberlist"
)

type MemberListOption func(ctx context.Context, config *memberlist.Config)

func MemberListOptionHost(host string) MemberListOption {
	return func(_ context.Context, config *memberlist.Config) {
		config.BindAddr = host
	}
}

func MemberListOptionPort(port int) MemberListOption {
	return func(_ context.Context, config *memberlist.Config) {
		config.BindPort = port
	}
}

// MemberList is a wrapper around hashicorp/memberlist to make it easier to
// utilize for integration into the cluster functionality.
type MemberList struct {
	memberList     *memberlist.Memberlist
	done           chan<- struct{}
	memberDelegate MemberDelegate
}

// Options can be provided in case it is desired to modify the base configuration
// used when creating the MemberList.
func NewMemberList(ctx context.Context, disco discovery.Discovery, metadata []byte, options ...MemberListOption) MemberList {
	memberDelegate := MemberDelegate{
		metadata: metadata,
	}

	// Create the initial memberlist from a safe configuration.
	// Please reference the godoc for other default config types.
	// http://godoc.org/github.com/hashicorp/memberlist#Config
	config := memberlist.DefaultLocalConfig()
	config.BindAddr = "0.0.0.0"
	config.BindPort = 7946
	config.Name = Uuid()
	config.Delegate = memberDelegate
	config.LogOutput = io.Discard
	config.Events = MemberEvents{
		OnJoin:   func(ctx context.Context, node *memberlist.Node) {},
		OnLeave:  func(ctx context.Context, node *memberlist.Node) {},
		OnUpdate: func(ctx context.Context, node *memberlist.Node) {},
	}

	for index := range options {
		options[index](ctx, config)
	}

	list, err := memberlist.Create(config)
	if err != nil {
		panic("failed to create memberlist: " + err.Error())
	}

	done := make(chan struct{})

	memberList := MemberList{
		done:           done,
		memberDelegate: memberDelegate,
		memberList:     list,
	}

	go memberList.discoverNodes(ctx, disco, done)

	return memberList
}

func (self MemberList) discoverNodes(ctx context.Context, disco discovery.Discovery, done <-chan struct{}) {
	tickPeriod := disco.DiscoveryPeriod()

	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()

	for {
		networkNodes := disco.Discover(ctx)

		// Shuffle the discovered nodes. This ensures that if the cluster ever gets split, it can eventually self-heal.
		// Consider a network of NodeA, NodeB, NodeC, NodeD. If NodeA and NodeB are clustered and NodeC and NodeD are clustered,
		// but both groups are not aware of the other, then the network becomes split and will never join. To fix this, whenever
		// a node discovers other nodes on the network, it will attempt a join with a random one. In this instance, eventually
		// one node from NodeA/NodeB or one node from NodeC/NodeD will eventually attempt a join a node in the other group. After
		// that occurs, the memberlist will propagate the cluster nodes via the gossip protocol, thus self-healing the cluster.
		rand.Shuffle(len(networkNodes), func(i, j int) { networkNodes[i], networkNodes[j] = networkNodes[j], networkNodes[i] })

		// fmt.Println("networkNodes", networkNodes)

		for index := range networkNodes {
			host := networkNodes[index].Host
			port := networkNodes[index].Port

			_, err := self.memberList.Join([]string{host + ":" + port})
			if err != nil {
				// fmt.Println("failed to join cluster: " + err.Error())
				// Keep trying until all network nodes are exhausted.
				continue
			}

			if len(self.memberList.Members()) == 1 && self.memberList.Members()[0].Name == self.Name() {
				// We have only connected to ourselves.
				continue
			}

			// We only need to join one node. The memberlist will connect this to the whole cluster eventually.
			// fmt.Printf("%s:%s joined network through node %s:%s\n", self.Host(), self.Port(), host, port)
			break
		}

		ticker.Reset(tickPeriod)

		select {
		case <-done:
			return
		case <-ticker.C:
			continue
		}
	}
}

func (self MemberList) Name() string {
	return self.memberList.LocalNode().Name
}

func (self MemberList) Host() string {
	return self.Node().Addr.String()
}

func (self MemberList) Port() string {
	return strconv.Itoa(int(self.Node().Port))
}

func (self MemberList) Shutdown(timeout time.Duration) error {
	self.done <- struct{}{}
	close(self.done)

	// Gracefully leave the cluster.
	if err := self.memberList.Leave(timeout); err != nil {
		return err
	}

	// Stop all background processes and appear "dead" to the cluster.
	return self.memberList.Shutdown()
}

func (self MemberList) Node() *memberlist.Node {
	return self.memberList.LocalNode()
}

func (self MemberList) Members() []*memberlist.Node {
	return self.memberList.Members()
}

func (self MemberList) MemberNames() []string {
	members := self.memberList.Members()
	names := make([]string, len(members))
	for index := range members {
		names[index] = members[index].Name
	}
	return names
}

func Uuid() string {
	uuid, err := uuid.GenerateUUID()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return uuid
}

type MemberDelegate struct {
	metadata []byte
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (self MemberDelegate) NodeMeta(limit int) []byte {
	// fmt.Printf("node meta: %d\n", limit)
	return self.metadata
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (self MemberDelegate) NotifyMsg(message []byte) {
	// fmt.Println("notify message: " + string(message))
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (self MemberDelegate) GetBroadcasts(overhead int, limit int) [][]byte {
	// fmt.Printf("get broadcasts: %d, %d\n", overhead, limit)
	return [][]byte{}
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (self MemberDelegate) LocalState(join bool) []byte {
	// fmt.Printf("local state: %t\n", join)
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (self MemberDelegate) MergeRemoteState(buf []byte, join bool) {
	// fmt.Printf("merge remote state: %s, %t\n", buf, join)
}

func MemberListOptionEventCallbacks(ctx context.Context, onJoin func(ctx context.Context, node *memberlist.Node), onLeave func(ctx context.Context, node *memberlist.Node), onUpdate func(ctx context.Context, node *memberlist.Node)) MemberListOption {
	return func(ctx context.Context, config *memberlist.Config) {
		config.Events = MemberEvents{
			ctx:      ctx,
			OnJoin:   onJoin,
			OnLeave:  onLeave,
			OnUpdate: onUpdate,
		}
	}
}

type MemberEvents struct {
	ctx      context.Context
	OnJoin   func(ctx context.Context, node *memberlist.Node)
	OnLeave  func(ctx context.Context, node *memberlist.Node)
	OnUpdate func(ctx context.Context, node *memberlist.Node)
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (self MemberEvents) NotifyJoin(node *memberlist.Node) {
	// fmt.Printf("node joined: %s %s\n", node.Name, node.Address())
	self.OnJoin(self.ctx, node)
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (self MemberEvents) NotifyLeave(node *memberlist.Node) {
	// fmt.Printf("node left: %s %s\n", node, node.Address())
	self.OnLeave(self.ctx, node)
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (self MemberEvents) NotifyUpdate(node *memberlist.Node) {
	// fmt.Printf("node updated: %s %s\n", node, node.Address())
	self.OnUpdate(self.ctx, node)
}
