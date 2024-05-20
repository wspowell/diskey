package cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type Value interface{}

func Get[T Value](cache Cache, key string) (T, error) {
	var value T

	valueBytes, err := cache.Get(key)
	if err != nil {
		return value, err
	}

	if err := UnmarshalValue(valueBytes, &value); err != nil {
		return value, err
	}

	return value, nil
}

func UnmarshalValue[T Value](valueBytes []byte, value *T) error {
	return msgpack.Unmarshal(valueBytes, &value)
}

func Set[T Value](cache Cache, key string, value T) error {
	valueBytes, err := MarshalValue(value)
	if err != nil {
		return err
	}
	return cache.Set(key, valueBytes)
}

func Delete(cache Cache, key string) error {
	return cache.cache.Delete(key)
}

func MarshalValue[T Value](value T) ([]byte, error) {
	valueBytes, err := msgpack.Marshal(value)
	if err != nil {
		return nil, err
	}
	return valueBytes, nil
}

type Config struct {
	OnKeyExpired func(key string, entry []byte)
	OnKeyEvicted func(key string, entry []byte)
	OnKeyDeleted func(key string, entry []byte)
}

type Cache struct {
	cache *bigcache.BigCache
}

func New(ctx context.Context, config Config) (Cache, error) {
	bigCacheConfig := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024, // Matches number of redis hash slots.

		// time after which entry can be evicted
		LifeWindow: 10 * time.Minute,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: 5 * time.Minute,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 10 * 10 * 60, // 1000 * 10 * 60,

		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 500,

		// prints information about additional memory allocation
		Verbose: false,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 1024,

		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,

		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: func(key string, entry []byte, reason bigcache.RemoveReason) {
			switch reason {
			case bigcache.Expired:
				if config.OnKeyExpired == nil {
					return
				}
				config.OnKeyExpired(key, entry)
			case bigcache.NoSpace:
				if config.OnKeyEvicted == nil {
					return
				}
				config.OnKeyEvicted(key, entry)
			case bigcache.Deleted:
				if config.OnKeyDeleted == nil {
					return
				}
				config.OnKeyDeleted(key, entry)
			}
		},
	}

	cache, err := bigcache.New(ctx, bigCacheConfig)
	if err != nil {
		return Cache{}, err
	}

	return Cache{
		cache: cache,
	}, nil
}

func (self Cache) Set(key string, value []byte) error {
	return self.cache.Set(key, value)
}

func (self Cache) Get(key string) ([]byte, error) {
	return self.cache.Get(key)
}

func (self Cache) Delete(key string) error {
	return self.cache.Delete(key)
}
