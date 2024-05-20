package cache_test

import (
	"context"
	"encoding/gob"
	"testing"

	"diskey/pkg/cache"

	"github.com/vmihailenco/msgpack/v5"
)

func Benchmark_Get(b *testing.B) {
	ctx := context.Background()
	cacheConfig := cache.Config{}
	storage, err := cache.New(ctx, cacheConfig)
	if err != nil {
		panic(err)
	}

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	if err := cache.Set(storage, "key", testValue); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.Get[MyValue](storage, "key")
		if err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func Benchmark_Set(b *testing.B) {
	ctx := context.Background()
	cacheConfig := cache.Config{}
	storage, err := cache.New(ctx, cacheConfig)
	if err != nil {
		panic(err)
	}

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cache.Set(storage, "key", testValue); err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

// goos: linux
// goarch: amd64
// pkg: diskey/pkg/cache
// cpu: AMD Ryzen 9 4900HS with Radeon Graphics
// Benchmark_vmihailenco_msgpack_Marshal-8     	 2981389	       362.7 ns/op	     136 B/op	       3 allocs/op
// Benchmark_vmihailenco_msgpack_Unmarshal-8   	 3081426	       379.2 ns/op	      76 B/op	       3 allocs/op
// Benchmark_golang_gob_Marshal-8              	  496046	      2280 ns/op	    1224 B/op	      22 allocs/op
// Benchmark_golang_gob_Unmarshal-8            	   66387	     17276 ns/op	    6744 B/op	     178 allocs/op

func Benchmark_vmihailenco_msgpack_Marshal(b *testing.B) {
	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msgpack.Marshal(testValue)
		if err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func Benchmark_vmihailenco_msgpack_Unmarshal(b *testing.B) {
	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	bytes, err := msgpack.Marshal(testValue)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var item MyValue
		err = msgpack.Unmarshal(bytes, &item)
		if err != nil {
			panic(err)
		}

	}
	b.StopTimer()
}

func Benchmark_golang_gob_Marshal(b *testing.B) {
	gob.Register(MyValue{})
	s := GobSerializer{}

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Marshal(testValue)
		if err != nil {
			panic(err)
		}
	}
	b.StopTimer()
}

func Benchmark_golang_gob_Unmarshal(b *testing.B) {
	gob.Register(MyValue{})
	s := GobSerializer{}

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	bytes, err := s.Marshal(testValue)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var item MyValue
		err = s.Unmarshal(bytes, &item)
		if err != nil {
			panic(err)
		}

	}
	b.StopTimer()
}
