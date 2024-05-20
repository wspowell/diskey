package cluster

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache/v3"

	"diskey/pkg/cache"
	"diskey/pkg/command"
)

type ClusterCommandRpcHandlers struct {
	*Cluster
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	ValueBytes []byte
	Exists     bool
}

func (self ClusterCommandRpcHandlers) Get(args GetArgs, reply *GetReply) error {
	valueBytes, err := self.keyStore.Get(args.Key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil
		}
		return err // FIXME: generic error
	}

	reply.ValueBytes = valueBytes
	reply.Exists = true

	return nil
}

func newGetRequest(key string, resp *GetReply) command.Request {
	return command.Request{
		Name:  "ClusterCommandRpcHandlers.Get",
		Args:  GetArgs{Key: key},
		Reply: resp,
	}
}

func Get[T cache.Value](cluster *Cluster, key string) (T, bool) {
	ownerAddress := cluster.getClosestAddress(key)
	if cluster.clusterServer.Address() == ownerAddress.String() {
		value, err := cache.Get[T](cluster.keyStore, key)
		if err != nil {
			if errors.Is(err, bigcache.ErrEntryNotFound) {
				return value, false
			}
			return value, false // FIXME: generic error
		}
		return value, true
	}

	response := &GetReply{}
	request := &keyRequest{
		key:     key,
		request: newGetRequest(key, response),
		done:    atomic.Bool{},
	}

	sendRequest(cluster, request)

	var value T
	if err := cache.UnmarshalValue(response.ValueBytes, &value); err != nil {
		return value, false
	}

	return value, response.Exists
}

type SetArgs struct {
	Key        string
	ValueBytes []byte
}

type SetReply struct{}

func (self ClusterCommandRpcHandlers) Set(args SetArgs, reply *SetReply) error {
	return self.keyStore.Set(args.Key, args.ValueBytes)
}

func newSetRequest(key string, valueBytes []byte, resp *SetReply) command.Request {
	return command.Request{
		Name: "ClusterCommandRpcHandlers.Set",
		Args: SetArgs{
			Key:        key,
			ValueBytes: valueBytes,
		},
		Reply: resp,
	}
}

func Set[T cache.Value](cluster *Cluster, key string, value T) error {
	ownerAddress := cluster.getClosestAddress(key)
	if cluster.clusterServer.Address() == ownerAddress.String() {
		return cache.Set(cluster.keyStore, key, value)
	}

	valueBytes, err := cache.MarshalValue(value)
	if err != nil {
		return err
	}

	response := &SetReply{}
	request := &keyRequest{
		key:     key,
		request: newSetRequest(key, valueBytes, response),
		done:    atomic.Bool{},
	}

	sendRequest(cluster, request)

	return nil
}

type DeleteArgs struct {
	Key string
}

type DeleteReply struct{}

func (self ClusterCommandRpcHandlers) Delete(args DeleteArgs, reply *DeleteReply) error {
	err := self.keyStore.Delete(args.Key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			// Not found errors on delete are not an error. This is the desired case.
			return nil
		}
		return err // FIXME: generic error
	}

	return nil
}

func newDeleteRequest(key string, resp *DeleteReply) command.Request {
	return command.Request{
		Name:  "ClusterCommandRpcHandlers.Delete",
		Args:  GetArgs{Key: key},
		Reply: resp,
	}
}

func Delete(cluster *Cluster, key string) error {
	ownerAddress := cluster.getClosestAddress(key)
	if cluster.clusterServer.Address() == ownerAddress.String() {
		err := cache.Delete(cluster.keyStore, key)
		if err != nil {
			if errors.Is(err, bigcache.ErrEntryNotFound) {
				return nil
			}
			return err // FIXME: generic error
		}
		return nil
	}

	response := &DeleteReply{}
	request := &keyRequest{
		key:     key,
		request: newDeleteRequest(key, response),
		done:    atomic.Bool{},
	}

	sendRequest(cluster, request)

	return nil
}

type BatchArgs struct {
	Requests []*command.Request
}

type BatchReply struct {
	Responses []any
}

func (self ClusterCommandRpcHandlers) Batch(args BatchArgs, reply *BatchReply) error {
	reply.Responses = make([]any, len(args.Requests))
	for index := range args.Requests {
		if err := runBatchRequest(self, args.Requests[index].Name, args.Requests[index].Args, &reply.Responses[index]); err != nil {
			return err
		}
	}

	return nil
}

func newBatchRequest(requests []*command.Request, responses *BatchReply) command.Request {
	return command.Request{
		Name: "ClusterCommandRpcHandlers.Batch",
		Args: BatchArgs{
			Requests: requests,
		},
		Reply: responses,
	}
}

type keyRequest struct {
	key     string
	request command.Request
	done    atomic.Bool
}

func runBatchRequest(handlers ClusterCommandRpcHandlers, name string, args any, reply *any) error {
	switch name {
	case "ClusterCommandRpcHandlers.Get":
		args := GetArgs{
			Key: args.(map[string]any)["Key"].(string),
		}
		getReply := &GetReply{}
		if err := handlers.Get(args, getReply); err != nil {
			return err
		}
		*reply = getReply
	case "ClusterCommandRpcHandlers.Set":
		args := SetArgs{
			Key:        args.(map[string]any)["Key"].(string),
			ValueBytes: args.(map[string]any)["ValueBytes"].([]byte),
		}
		setReply := &SetReply{}
		if err := handlers.Set(args, setReply); err != nil {
			return err
		}
		*reply = setReply
	case "ClusterCommandRpcHandlers.Delete":
		args := DeleteArgs{
			Key: args.(map[string]any)["Key"].(string),
		}
		deleteReply := &DeleteReply{}
		if err := handlers.Delete(args, deleteReply); err != nil {
			return err
		}
		*reply = deleteReply
	}
	return nil
}

func (self *Cluster) runBatch(ctx context.Context, keyRequests []*keyRequest) error {
	requestsByClient := map[Address][]*command.Request{}

	handlers := ClusterCommandRpcHandlers{
		Cluster: self,
	}

	for index := range keyRequests {
		ownerAddress := self.getClosestAddress(keyRequests[index].key)

		if ownerAddress.String() == self.clusterServer.Address() {
			if err := runBatchRequest(handlers, keyRequests[index].request.Name, keyRequests[index].request.Args, &keyRequests[index].request.Reply); err != nil {
				return err
			}
		} else {
			requestsByClient[ownerAddress] = append(requestsByClient[ownerAddress], &keyRequests[index].request)
		}
	}

	for address, requests := range requestsByClient {
		clientKeyOwner := self.getClientByHostPort(address.Host, address.Port)
		if clientKeyOwner != nil {
			response := &BatchReply{}
			request := newBatchRequest(requests, response)
			if err := clientKeyOwner.Send(ctx, request); err != nil {
				return err
			}

			for index := range requests {
				commandRequest := &(requests[index])

				switch requests[index].Name {
				case "ClusterCommandRpcHandlers.Get":
					valueBytes := response.Responses[index].(map[string]any)["ValueBytes"]
					if valueBytes != nil {
						(*commandRequest).Reply.(*GetReply).ValueBytes = valueBytes.([]byte)
					}
					(*commandRequest).Reply.(*GetReply).Exists = response.Responses[index].(map[string]any)["Exists"].(bool)
				case "ClusterCommandRpcHandlers.Set":
					// SetReply is an empty body.
				case "ClusterCommandRpcHandlers.Delete":
					// DeleteReply is an empty body.
				}
			}
		}
	}

	for index := range keyRequests {
		keyRequests[index].done.Store(true)
	}

	return nil
}

func sendRequest(cluster *Cluster, request *keyRequest) {
	cluster.batchChannel <- request

	deadline := time.Now().Add(5 * time.Second)
	for {
		if request.done.Load() {
			break
		}
		if time.Now().After(deadline) {
			break
		}
	}
}
