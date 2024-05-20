package discovery

import (
	"context"
	"time"
)

type DiscoveredNode struct {
	Host string
	Port string
}

type Discovery interface {
	// Discovery nodes on the network.
	// The network could be local, WAN, or any other topology.
	// This function return a host:port that can be used for connecting to the cluster.
	Discover(ctx context.Context) []DiscoveredNode
	DiscoveryPeriod() time.Duration
}
