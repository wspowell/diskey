package cluster

import (
	"diskey/pkg/cluster/internal/hashtag"
)

type HashSlot uint16

const MaxHashSlot HashSlot = 16384

func Slot(key string) HashSlot {
	return HashSlot(hashtag.Slot(key))
}

type Range struct {
	Begin HashSlot
	End   HashSlot
}

func (self Range) Contains(testSlot HashSlot) bool {
	return self.Begin <= testSlot && testSlot <= self.End
}

func (self Range) Overlaps(other Range) bool {
	return (other.Begin <= self.Begin && self.Begin <= other.End) ||
		(other.Begin <= self.End && self.End <= other.End) ||
		other.Overlaps(self)
}

// Super naive algorithm to get a first approximation at clustering for initial functionality.
//
// Hash each address to a slot, then hash the key and see which address slot the key is closest to and that is the owner.

type Address struct {
	Host string
	Port string
	Slot HashSlot
}

func (self Address) String() string {
	return self.Host + ":" + self.Port
}

func (self *Cluster) getClosestAddress(key string) Address {
	keySlot := int(Slot(key))

	var closestAddress Address
	var closestAddressSlot int
	var closestDistance int

	self.clientsMutex.RLock()
	defer self.clientsMutex.RUnlock()
	for index := range self.addresses {
		addressSlot := int(self.addresses[index].Slot)

		// Calculate the "wraparound" distance. For example 0 and 16384 are distance 1 from each other.
		distance := abs(keySlot - addressSlot)
		if distance > int(MaxHashSlot)/2 {
			distance = 1 - distance
		}

		if closestAddress.Host == "" && closestAddress.Port == "" {
			closestAddress = self.addresses[index]
			closestAddressSlot = addressSlot
			closestDistance = distance
			continue
		}

		if distance <= closestDistance {
			if distance == closestDistance {
				// Lowest address hash slot wins.
				if addressSlot < closestAddressSlot {
					closestAddress = self.addresses[index]
					closestAddressSlot = addressSlot
					closestDistance = distance
				}
				continue
			}
			closestAddress = self.addresses[index]
			closestAddressSlot = addressSlot
			closestDistance = distance
		}
	}

	return closestAddress
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}
