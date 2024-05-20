package cache_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"

	"diskey/pkg/cache"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

type MyValue struct {
	Foo int
	Bar string
}

func Test_Get_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheConfig := cache.Config{}
	storage, err := cache.New(ctx, cacheConfig)
	assert.NoError(t, err)

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	assert.NoError(t, cache.Set(storage, "key", testValue))
	actualValue, err := cache.Get[MyValue](storage, "key")
	assert.NoError(t, err)
	assert.Equal(t, testValue, actualValue)
}

func Test_vmihailenco_msgpack(t *testing.T) {
	t.Parallel()

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	b, err := msgpack.Marshal(testValue)
	if err != nil {
		panic(err)
	}

	var item MyValue
	err = msgpack.Unmarshal(b, &item)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, testValue, item)

	// gob.NewEncoder()
}

type GobSerializer struct{}

func (g *GobSerializer) Marshal(o interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(o)
	return buf.Bytes(), err
}

func (g *GobSerializer) Unmarshal(d []byte, o interface{}) error {
	return gob.NewDecoder(bytes.NewReader(d)).Decode(o)
}

func Test_gob(t *testing.T) {
	t.Parallel()

	testValue := MyValue{
		Foo: 10,
		Bar: "test",
	}

	gob.Register(MyValue{})
	s := GobSerializer{}

	b, err := s.Marshal(testValue)
	if err != nil {
		panic(err)
	}

	var item MyValue
	err = s.Unmarshal(b, &item)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, testValue, item)

	// gob.NewEncoder()
}
