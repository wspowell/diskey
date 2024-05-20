package cluster_test

import (
	"context"
	"testing"
	"time"

	"diskey/pkg/cluster"

	"github.com/stretchr/testify/assert"
)

func waitForCluster(caches ...*cluster.Cluster) {
	timeout := time.Tick(30 * time.Second)
	for {
		select {
		case <-timeout:
			panic("timed out waiting for cluster")
		default:
		}

		var waiting bool
		for index := range caches {
			if caches[index].NumClients() != len(caches)-1 {
				waiting = true
				break
			}
		}

		if waiting {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// fmt.Println("cluster ready")
		break
	}
}

type MyValue struct {
	Foo int
	Bar string
}

func TestCluster_Set_Get_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cache1 := cluster.NewCluster(ctx, "localhost", "7000", cluster.OptionMemberListPort("7950"), cluster.OptionLocalhostDiscovery([]string{"7950", "7951"}))
	cache2 := cluster.NewCluster(ctx, "localhost", "7001", cluster.OptionMemberListPort("7951"), cluster.OptionLocalhostDiscovery([]string{"7950", "7951"}))
	waitForCluster(cache1, cache2)

	value, exists := cluster.Get[MyValue](cache1, "key")
	assert.Equal(t, MyValue{}, value)
	assert.Equal(t, false, exists)

	value, exists = cluster.Get[MyValue](cache2, "key")
	assert.Equal(t, MyValue{}, value)
	assert.Equal(t, false, exists)

	// Sets on cache1 first
	{
		expectedValue := MyValue{
			Foo: 10,
			Bar: "test1",
		}

		err := cluster.Set(cache1, "key", expectedValue)
		assert.NoError(t, err)

		value, exists := cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)
	}

	// Sets on cache2 first.
	{
		expectedValue := MyValue{
			Foo: 15,
			Bar: "test2",
		}

		err := cluster.Set(cache2, "key", expectedValue)
		assert.NoError(t, err)

		value, exists := cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)
	}
}

func TestCluster_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cache1 := cluster.NewCluster(ctx, "localhost", "7002", cluster.OptionMemberListPort("8000"), cluster.OptionLocalhostDiscovery([]string{"8000", "8001"}))
	cache2 := cluster.NewCluster(ctx, "localhost", "7003", cluster.OptionMemberListPort("8001"), cluster.OptionLocalhostDiscovery([]string{"8000", "8001"}))
	waitForCluster(cache1, cache2)

	expectedValue := MyValue{
		Foo: 10,
		Bar: "test1",
	}

	// Deletes on cache1.
	{
		err := cluster.Set(cache1, "key", expectedValue)
		assert.NoError(t, err)

		value, exists := cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		assert.NoError(t, cluster.Delete(cache1, "key"))

		value, exists = cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, MyValue{}, value)
		assert.Equal(t, false, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, MyValue{}, value)
		assert.Equal(t, false, exists)
	}

	// Deletes on cache2.
	{
		err := cluster.Set(cache1, "key", expectedValue)
		assert.NoError(t, err)

		value, exists := cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, exists)

		assert.NoError(t, cluster.Delete(cache2, "key"))

		value, exists = cluster.Get[MyValue](cache1, "key")
		assert.Equal(t, MyValue{}, value)
		assert.Equal(t, false, exists)

		value, exists = cluster.Get[MyValue](cache2, "key")
		assert.Equal(t, MyValue{}, value)
		assert.Equal(t, false, exists)
	}
}
