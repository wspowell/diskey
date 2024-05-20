package cluster_test

import (
	"context"
	"testing"
	"time"

	"diskey/pkg/cluster"
	"diskey/pkg/discovery"

	"github.com/stretchr/testify/assert"
)

// func Test_Cluster(t *testing.T) {
// 	t.Parallel()

// 	ctx := log.Logger.Level(zerolog.DebugLevel).WithContext(context.Background())

// 	cluster8000 := cluster.NewCluster(ctx, "localhost", "8000", cluster.OptionLocalhostReloader([]string{"8001"}))
// 	cluster8001 := cluster.NewCluster(ctx, "localhost", "8001", cluster.OptionLocalhostReloader([]string{"8000"}))

// 	cluster8000.Reload(ctx)
// 	cluster8001.Reload(ctx)

// 	assert.NoError(t, cluster8000.Ping(ctx))
// 	assert.NoError(t, cluster8001.Ping(ctx))

// 	t.Fail()
// }

func Test_MemberList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	localhostNetwork := discovery.NewLocalhost([]string{"7960", "7961", "7962"})

	list1 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7960))
	list2 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7961))
	list3 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7962))

	// Wait for each member to get updated with the whole cluster.
	for {
		if len(list1.Members()) == 3 && len(list2.Members()) == 3 && len(list3.Members()) == 3 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	assert.ElementsMatch(t, list1.MemberNames(), list2.MemberNames())
	assert.ElementsMatch(t, list1.MemberNames(), list3.MemberNames())
}

func Test_MemberList_shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	localhostNetwork := discovery.NewLocalhost([]string{"7970", "7971", "7972"})

	list1 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7970))
	list2 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7971))
	list3 := cluster.NewMemberList(ctx, localhostNetwork, nil, cluster.MemberListOptionPort(7972))

	for {
		if len(list1.Members()) == 3 && len(list2.Members()) == 3 && len(list3.Members()) == 3 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	assert.NoError(t, list1.Shutdown(time.Second))

	for {
		if len(list1.Members()) == 2 && len(list2.Members()) == 2 && len(list3.Members()) == 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	assert.NotContains(t, list1.MemberNames(), list1.Name())
	assert.NotContains(t, list2.MemberNames(), list1.Name())
	assert.NotContains(t, list3.MemberNames(), list1.Name())
}

func Test_MemberList_split_cluster_self_heal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	localhostNetwork1 := discovery.NewLocalhost([]string{"7980", "7981"})
	localhostNetwork1.Period = time.Second
	localhostNetwork2 := discovery.NewLocalhost([]string{"7982", "7983"})
	localhostNetwork2.Period = time.Second

	list1 := cluster.NewMemberList(ctx, localhostNetwork1, nil, cluster.MemberListOptionPort(7980))
	list2 := cluster.NewMemberList(ctx, localhostNetwork1, nil, cluster.MemberListOptionPort(7981))
	list3 := cluster.NewMemberList(ctx, localhostNetwork2, nil, cluster.MemberListOptionPort(7982))
	list4 := cluster.NewMemberList(ctx, localhostNetwork2, nil, cluster.MemberListOptionPort(7983))

	// Wait for each member to get updated with the whole cluster.
	for {
		if len(list1.Members()) == 2 && len(list2.Members()) == 2 && len(list3.Members()) == 2 && len(list4.Members()) == 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// Tell the first network about one of the other nodes.
	localhostNetwork1.AddPort("7982")

	// Wait for each member to get updated with the whole cluster.
	for {
		if len(list1.Members()) == 4 && len(list2.Members()) == 4 && len(list3.Members()) == 4 && len(list4.Members()) == 4 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// Even though the last node has no way of connecting to the other network, and vice versa, the gossip chain notifies
	// it of the other nodes, and vice versa, therefore healing the network and making all four nodes aware of each other.
	assert.ElementsMatch(t, list1.MemberNames(), list2.MemberNames())
	assert.ElementsMatch(t, list1.MemberNames(), list3.MemberNames())
	assert.ElementsMatch(t, list1.MemberNames(), list4.MemberNames())
}

// func Test_Memberlist(t *testing.T) {
// 	t.Parallel()

// 	/* Create the initial memberlist from a safe configuration.
// 	   Please reference the godoc for other default config types.
// 	   http://godoc.org/github.com/hashicorp/memberlist#Config
// 	*/
// 	config1 := memberlist.DefaultLocalConfig()
// 	config1.BindPort = 7946
// 	config1.Name = cluster.Uuid()
// 	list1, err := memberlist.Create(config1)
// 	if err != nil {
// 		panic("Failed to create memberlist: " + err.Error())
// 	}

// 	// Ask for members of the cluster
// 	for _, member := range list1.Members() {
// 		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
// 	}

// 	// Join an existing cluster by specifying at least one known member.
// 	n, err := list1.Join([]string{"172.24.0.2:7946"})
// 	if err != nil {
// 		panic("Failed to join cluster: " + err.Error())
// 	}
// 	fmt.Println(n)

// 	n, err = list1.Join([]string{"172.24.0.2"})
// 	if err != nil {
// 		panic("Failed to join cluster: " + err.Error())
// 	}
// 	fmt.Println(n)

// 	// Ask for members of the cluster
// 	for _, member := range list1.Members() {
// 		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
// 	}

// 	config2 := memberlist.DefaultLocalConfig()
// 	config2.BindPort = 7947
// 	config2.Name = cluster.Uuid()
// 	config2.Delegate = cluster.Node{}
// 	list2, err := memberlist.Create(config2)
// 	if err != nil {
// 		panic("Failed to create memberlist: " + err.Error())
// 	}

// 	// Ask for members of the cluster
// 	for _, member := range list2.Members() {
// 		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
// 	}

// 	n, err = list1.Join([]string{"172.24.0.2:7947"})
// 	if err != nil {
// 		panic("Failed to join cluster: " + err.Error())
// 	}
// 	fmt.Println(n)

// 	// Ask for members of the cluster
// 	for _, member := range list2.Members() {
// 		fmt.Printf("Member: %s %s %d\n", member.Name, member.Addr, member.Port)
// 		if err := list1.SendReliable(member, []byte(`hello`)); err != nil {
// 			panic("Failed to send message: " + err.Error())
// 		}
// 		if err := list1.SendBestEffort(member, []byte(`ohai`)); err != nil {
// 			panic("Failed to send message: " + err.Error())
// 		}
// 	}

// 	// Continue doing whatever you need, memberlist will maintain membership
// 	// information in the background. Delegates can be used for receiving
// 	// events when members join or leave.

// 	time.Sleep(1 * time.Second)

// 	t.Fail()
// }
