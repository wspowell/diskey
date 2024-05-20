package diskey

import (
	"context"

	"diskey/pkg/cache"
	"diskey/pkg/cluster"
	"diskey/pkg/discovery"
)

type clientContextKey struct{}

type Config struct {
	Host               string
	ServerToServerPort string
	MemberListPort     string
}

type Client struct {
	diskeyCluster *cluster.Cluster
}

func NewClient(ctx context.Context, config Config, disco discovery.Discovery) Client {
	return Client{
		diskeyCluster: cluster.NewCluster(
			ctx,
			config.Host,
			config.ServerToServerPort,
			cluster.OptionDiscovery(disco),
			cluster.OptionMemberListPort(config.MemberListPort),
		),
	}
}

func WithContext(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientContextKey{}, client)
}

func fromContext(ctx context.Context) Client {
	if client, ok := ctx.Value(clientContextKey{}).(Client); ok {
		return client
	}
	panic("WithContext() was never called to add the Client to the Context")
}

func Get[T cache.Value](ctx context.Context, key string) (T, bool) {
	return cluster.Get[T](fromContext(ctx).diskeyCluster, key)
}

func Set[T cache.Value](ctx context.Context, key string, value T) {
	_ = cluster.Set[T](fromContext(ctx).diskeyCluster, key, value)
}

func Delete(ctx context.Context, key string) {
	_ = cluster.Delete(fromContext(ctx).diskeyCluster, key)
}
