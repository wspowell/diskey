package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
	// "diskey/pkg/cluster"
	"diskey/pkg/discovery"
	"diskey/pkg/diskey"
)

func runClient(ctx context.Context, serverPort string) {
	// diskeyConfig := diskey.Config{
	// 	ClusterConfig: cluster.Config{
	// 		ServerPort: serverPort,
	// 	},
	// }
	// diskeyClient, clientErr := diskey.NewClient(ctx, diskeyConfig)
	// if clientErr.IsErr() {
	// 	log.Ctx(ctx).Err(clientErr).Send()
	// 	return
	// }

	// time.Sleep(time.Second)

	// value, _ := diskeyClient.Get(ctx, "test")
	// log.Ctx(ctx).Debug().Msg(string(value))

	// ctx = log.Ctx(ctx).With().
	// 	Str("originServerPort", serverPort).
	// 	Logger().WithContext(ctx)

	// clusterConfig := cluster.Config{
	// 	ServerPort: serverPort,
	// }
	// clusterClient := cluster.New(clusterConfig)
	// if err := clusterClient.Listen(ctx); err.IsErr() {
	// 	log.Ctx(ctx).Err(err)
	// 	return
	// } else {
	// 	log.Ctx(ctx).Debug().Err(err).Send()
	// }

	// for index := range clusterClients {
	// 	if err := clusterClient.RegisterNode(ctx, "localhost", clusterClients[index]); err.IsErr() {
	// 		log.Ctx(ctx).Err(err)
	// 		return
	// 	}
	// }

	// clusterClient.ForEachClient(ctx, func(ctx context.Context, client *client.Client) {
	// 	if client.Ping(ctx) {
	// 		log.Ctx(ctx).Debug().Msgf("pinged client: %s", client.Address())
	// 	} else {
	// 		log.Ctx(ctx).Error().Msgf("failed to ping client: %s", client.Address())
	// 	}
	// })
}

// This is a test main for now. Eventually it will a CLI tool
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logLevel := zerolog.DebugLevel
	zerolog.SetGlobalLevel(logLevel)
	ctx = zerolog.New(os.Stdout).Level(logLevel).WithContext(ctx)

	localhostDiscovery := discovery.NewLocalhost([]string{"7950", "7951"})

	go client1(ctx, localhostDiscovery)
	go client2(ctx, localhostDiscovery)

	for {
		time.Sleep(time.Millisecond)
	}

	// go runClient(ctx, "60000")
	// go runClient(ctx, "60001")

	// time.Sleep(1 * time.Hour)

	// cancel()

	// time.Sleep(5 * time.Second)
}

func client1(ctx context.Context, disco discovery.Discovery) {
	config := diskey.Config{
		Host:               "0.0.0.0",
		ServerToServerPort: "8000",
		MemberListPort:     "7950",
	}

	client := diskey.NewClient(ctx, config, disco)
	ctx = diskey.WithContext(ctx, client)

	type Foo struct {
		Test string
	}

	for {
		foo, exists := diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		diskey.Set(ctx, "key", Foo{
			Test: "hello, world!",
		})

		foo, exists = diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		diskey.Delete(ctx, "key")

		foo, exists = diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		time.Sleep(1 * time.Second)
	}
}

func client2(ctx context.Context, disco discovery.Discovery) {
	config := diskey.Config{
		Host:               "localhost",
		ServerToServerPort: "8001",
		MemberListPort:     "7951",
	}

	client := diskey.NewClient(ctx, config, disco)
	ctx = diskey.WithContext(ctx, client)

	type Foo struct {
		Test string
	}

	for {
		foo, exists := diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		diskey.Set(ctx, "key", Foo{
			Test: "hello, world!",
		})

		foo, exists = diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		diskey.Delete(ctx, "key")

		foo, exists = diskey.Get[Foo](ctx, "key")
		fmt.Println(foo, exists)

		time.Sleep(5 * time.Second)
	}
}
