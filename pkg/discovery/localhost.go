package discovery

import (
	"context"
	"sync"
	"time"
)

type Localhost struct {
	Host   string
	Ports  []string
	Period time.Duration
	mutex  sync.RWMutex
}

func NewLocalhost(ports []string) *Localhost {
	return &Localhost{
		Host:   "localhost",
		Ports:  ports,
		Period: time.Second, // time.Minute,
		mutex:  sync.RWMutex{},
	}
}

func (self *Localhost) DiscoveryPeriod() time.Duration {
	return self.Period
}

func (self *Localhost) Discover(ctx context.Context) []DiscoveredNode {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	nodes := make([]DiscoveredNode, len(self.Ports))
	for index := range self.Ports {
		nodes[index] = DiscoveredNode{
			Host: self.Host,
			Port: self.Ports[index],
		}
	}
	return nodes
}

func (self *Localhost) AddPort(port string) {
	self.mutex.Lock()
	self.Ports = append(self.Ports, port)
	self.mutex.Unlock()
}
