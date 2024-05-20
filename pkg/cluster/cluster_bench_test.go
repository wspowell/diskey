package cluster_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"diskey/pkg/cluster"
)

// 2024-05-05
//
// goos: linux
// goarch: amd64
// pkg: diskey/pkg/cluster
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// BenchmarkCache_Get_local-8              	45544782	        26.62 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCache_Get_remote-8             	   10000	    133094 ns/op	    2239 B/op	      29 allocs/op
// BenchmarkCache_Set_local-8              	    7948	    132208 ns/op	    2332 B/op	      26 allocs/op
// BenchmarkCache_Set_remote-8             	    8964	    129639 ns/op	    2303 B/op	      26 allocs/op
// BenchmarkCache_Get_local_100_nodes-8   	44747506	        27.76 ns/op	       3 B/op	       0 allocs/op
// BenchmarkCache_Get_remote_100_nodes-8   	   12576	    131533 ns/op	   17379 B/op	      62 allocs/op
// BenchmarkCache_Set_local_100_nodes-8   	      91	  14265362 ns/op	 1820518 B/op	    6294 allocs/op
// BenchmarkCache_Set_remote_100_nodes-8   	      88	  11749921 ns/op	 6962450 B/op	   16393 allocs/op

// 2024-05-15
//
// goos: linux
// goarch: amd64
// pkg: diskey/pkg/cluster
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// BenchmarkCache_Get_local-8    	 1877566	       635.1 ns/op	     105 B/op	       6 allocs/op
// BenchmarkCache_Get_remote-8   	    7935	    146067 ns/op	    2949 B/op	      32 allocs/op
// BenchmarkCache_Set_local-8    	 1215291	      1011 ns/op	     851 B/op	       7 allocs/op
// BenchmarkCache_Set_remote-8   	   19687	    135015 ns/op	    3686 B/op	      31 allocs/op

// 2024-05-19
//
// goos: linux
// goarch: amd64
// pkg: diskey/pkg/cluster
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// BenchmarkCache_Get_local-8             	 1885249	       626.5 ns/op	     105 B/op	       6 allocs/op
// BenchmarkCache_Get_remote-8            	    7341	    169630 ns/op	    2472 B/op	      51 allocs/op
// BenchmarkCache_Get_remote_parallel-8   	     100	  18131248 ns/op	   22588 B/op	      56 allocs/op
// BenchmarkCache_Set_local-8             	 1674886	       666.1 ns/op	     154 B/op	       3 allocs/op
// BenchmarkCache_Set_remote-8            	    5901	    579894 ns/op	    2453 B/op	      42 allocs/op
// BenchmarkCache_Delete_local-8          	  265212	      4607 ns/op	      98 B/op	       2 allocs/op
// BenchmarkCache_Delete_remote-8         	      19	  88062463 ns/op	  117168 B/op	     167 allocs/op
// BenchmarkCache_Get_local_100_nodes-8   	    6141	    213379 ns/op	   25543 B/op	     106 allocs/op
// BenchmarkCache_Get_remote_100_nodes-8   	    5877	    223349 ns/op	   40912 B/op	     149 allocs/op
// BenchmarkCache_Set_local_100_nodes-8   	    6828	    333008 ns/op	   68893 B/op	     194 allocs/op
// BenchmarkCache_Set_remote_100_nodes-8   	    5672	    587528 ns/op	   75740 B/op	     229 allocs/op

type BenchValue struct {
	Foo int
	Bar string
}

func getCluster(ctx context.Context, numNodes int) []*cluster.Cluster {
	time.Sleep(time.Second) // FIXME(hacky): Wait for the previous test to shutdown its server.

	caches := make([]*cluster.Cluster, numNodes)

	port := 7000
	memberPorts := make([]string, numNodes)
	serverPorts := make([]string, numNodes)
	for i := range numNodes {
		memberPorts[i] = strconv.Itoa(port)
		port++
		serverPorts[i] = strconv.Itoa(port)
		port++
	}

	for i := range numNodes {
		caches[i] = cluster.NewCluster(ctx, "localhost", serverPorts[i], cluster.OptionMemberListPort(memberPorts[i]), cluster.OptionLocalhostDiscovery(memberPorts))
	}

	waitForCluster(caches...)

	return caches
}

func shutdownCluster(caches ...*cluster.Cluster) {
	for index := range caches {
		caches[index].Close()
	}
}

func BenchmarkCache_Get_local(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotValue, exists := cluster.Get[BenchValue](caches[0], "key")
		if gotValue != value || !exists {
			panic("unexpected result")
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Get_remote(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotValue, exists := cluster.Get[BenchValue](caches[1], "key")
		if gotValue != value || !exists {
			panic("unexpected result")
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Get_remote_parallel(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gotValue, exists := cluster.Get[BenchValue](caches[1], "key")
			if gotValue != value || !exists {
				panic(fmt.Sprintf("unexpected result: %+v, %t", gotValue, exists))
			}
		}
	})
	b.StopTimer()
}

func BenchmarkCache_Set_local(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cluster.Set(caches[0], "key", value); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Set_remote(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cluster.Set(caches[1], "key", value); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Delete_local(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := cluster.Set(caches[0], "key", value); err != nil {
			panic(err)
		}
		b.StartTimer()

		if err := cluster.Delete(caches[0], "key"); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Delete_remote(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 2)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := cluster.Set(caches[0], "key", value); err != nil {
			panic(err)
		}
		b.StartTimer()

		if err := cluster.Delete(caches[1], "key"); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Get_local_100_nodes(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 100)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotValue, exists := cluster.Get[BenchValue](caches[0], "key")
		if gotValue != value || !exists {
			panic("unexpected result")
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Get_remote_100_nodes(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 100)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotValue, exists := cluster.Get[BenchValue](caches[1], "key")
		if gotValue != value || !exists {
			panic("unexpected result")
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Set_local_100_nodes(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 100)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cluster.Set(caches[0], "key", value); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func BenchmarkCache_Set_remote_100_nodes(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	caches := getCluster(ctx, 100)
	defer shutdownCluster(caches...)

	value := BenchValue{
		Foo: 10,
		Bar: "bench",
	}

	if err := cluster.Set(caches[0], "key", value); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cluster.Set(caches[1], "key", value); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}
