package cluster_test

import (
	"testing"

	"diskey/pkg/cluster"

	"github.com/stretchr/testify/assert"
)

func Test_Slot(t *testing.T) {
	t.Parallel()

	// These should never change between runs.
	assert.Equal(t, cluster.HashSlot(0x281), cluster.Slot("test0"))
	assert.Equal(t, cluster.HashSlot(0x12a0), cluster.Slot("test1"))
	assert.Equal(t, cluster.HashSlot(0x22c3), cluster.Slot("test2"))
	assert.Equal(t, cluster.HashSlot(0x32e2), cluster.Slot("test3"))
	assert.Equal(t, cluster.HashSlot(0x205), cluster.Slot("test4"))
	assert.Equal(t, cluster.HashSlot(0x1224), cluster.Slot("test5"))
	assert.Equal(t, cluster.HashSlot(0x2247), cluster.Slot("test6"))
	assert.Equal(t, cluster.HashSlot(0x3266), cluster.Slot("test7"))
	assert.Equal(t, cluster.HashSlot(0x389), cluster.Slot("test8"))
	assert.Equal(t, cluster.HashSlot(0x13a8), cluster.Slot("test9"))
}
