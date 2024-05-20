package rpc_test

import (
	"context"
	"errors"
	"log"
	"net"
	gorpc "net/rpc"
	"sync"
	"testing"

	"diskey/pkg/command"
	"diskey/pkg/rpc"
)

// Codec: gob
// goos: linux
// goarch: amd64
// pkg: diskey/pkg/server
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// BenchmarkTest-8            	    7822	    145434 ns/op	     502 B/op	      16 allocs/op
// BenchmarkPing-8            	    7725	    157496 ns/op	    2480 B/op	      24 allocs/op
// BenchmarkPing_parallel-8   	   35712	     34134 ns/op	    2478 B/op	      24 allocs/op

// Codec: msgpack
// goos: linux
// goarch: amd64
// pkg: diskey/pkg/server
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// BenchmarkTest-8            	    8336	    140708 ns/op	     501 B/op	      16 allocs/op
// BenchmarkPing-8            	    8246	    144255 ns/op	    2416 B/op	      20 allocs/op
// BenchmarkPing_parallel-8   	   35959	     32518 ns/op	    2424 B/op	      20 allocs/op

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

var registerOnce = &sync.Once{}

func BenchmarkTest(b *testing.B) {
	arith := new(Arith)

	registerOnce.Do(func() {
		gorpc.Register(arith)
	})

	l, err := net.Listen("tcp", "localhost:1236")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()
	go func() {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		gorpc.ServeConn(conn)
	}()

	client, err := gorpc.Dial("tcp", "localhost:1236")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Synchronous call
		args := &Args{7, 8}
		var reply int
		err = client.Call("Arith.Multiply", args, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	}
	b.StopTimer()
	// fmt.Printf("Arith: %d*%d=%d", args.A, args.B, reply)
}

var (
	testServer rpc.Server
	testClient *rpc.Client
)

func init() {
	ctx := context.Background()
	port := "7100"

	testServer = rpc.NewServer("localhost", port)

	listener, listenErr := testServer.Listen(ctx)
	if listenErr.IsErr() {
		panic(listenErr)
	}

	go testServer.AcceptConnections(ctx, listener)

	testClient = rpc.NewClient("localhost", port)
	if connectErr := testClient.Connect(ctx); connectErr.IsErr() {
		panic(connectErr)
	}
}

func BenchmarkPing(b *testing.B) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	ctx := context.Background()

	// // Start CPU profiling
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	if testClient == nil {
		panic("client is nil")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		request := command.NewPingRequest()
		err := testClient.Send(ctx, request)
		if err != nil {
			panic(err)
		}

		if result, ok := request.Reply.(*int); !ok || *result != 1 {
			panic("ping failed")
		}
	}

	b.StopTimer()
}

func BenchmarkPing_parallel(b *testing.B) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	ctx := context.Background()

	// // Start CPU profiling
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			request := command.NewPingRequest()
			err := testClient.Send(ctx, request)
			if err != nil {
				panic(err)
			}

			if result, ok := request.Reply.(*int); !ok || *result != 1 {
				panic("ping failed")
			}
		}
	})

	b.StopTimer()
}
