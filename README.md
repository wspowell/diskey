# diskey

Which key? Dis key. Dis one.

This library provides a distributed in-memory key/value store that shards keys across a cluster of processes.

## Usage

The interface is generic and all (un)marshalling done for server to server communication is handled behind the scenes. A major goal is to keep the interface and usage as simple as possible.

### Discovery of Nodes

While the clustering of nodes is handled automatically, the library needs to know how and where to look for other nodes. This is uses a Discovery interface to provide that information. Currently, there is a localhost discovery implementation where you provide the ports to look for and when a node connects on one of those ports, it will be added to the cluster.

For example:
```
disco := discovery.NewLocalhost([]string{"7950", "7951"})
```

### Configure

Setting up the client requires telling it what port to communicate on for distributing keys and another for which it will coordinate clustering.
```
config := diskey.Config{
    Host:               "0.0.0.0",
    ServerToServerPort: "8000",
    MemberListPort:     "7950",
}

client := diskey.NewClient(ctx, config, disco)
ctx = diskey.WithContext(ctx, client)
```

The `WithContext()` function provides a convenient way of passing your client to all functions in your application for easy access. The API functions look for a client in the `Context` to use.

### diskey API

Assume you have some type you want to cache:
```
type Foo struct {
    Test string
}
```

Get a key:
```
foo, exists := diskey.Get[Foo](ctx, "key")
```

Set a key:
```
diskey.Set(ctx, "key", Foo{
    Test: "hello, world!",
})
```

Delete a key:
```
diskey.Delete(ctx, "key")
```

The API functions do not return errors because it not interesting or useful. We simply want to get, set, and delete keys, so the data is either there or it is not.