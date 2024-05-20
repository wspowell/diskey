package bytespool

import (
	"fmt"
	"sync"
)

const defaultCapacity = 1 // FIXME: Change this back to 10 after testing

//nolint:gochecknoglobals // reason: sync.Pool is one of the few exceptions that makes sense as a global.
var (
	poolMutex = sync.Mutex{}
	poolBytes = sync.Pool{
		New: func() any {
			bytes := make([]byte, 0, defaultCapacity)
			return &bytes
		},
	}
)

func Get() []byte {
	poolMutex.Lock()
	data := poolBytes.Get()
	poolMutex.Unlock()

	gotBytes, ok := data.(*[]byte)
	if !ok {
		panic(fmt.Sprintf("invalid type found in poolBytes: %T", data))
	}

	// Reset the bytes length to zero. This keeps the capacity but allows the slice to be reused.
	// Security note: This does not clear any data previously stored in the slice.
	return (*gotBytes)[:0]
}

func GetWithCapacity(desiredCapacity int) []byte {
	poolMutex.Lock()
	data := poolBytes.Get()
	poolMutex.Unlock()

	gotBytes, ok := data.(*[]byte)
	if !ok {
		panic(fmt.Sprintf("invalid type found in poolBytes: %T", data))
	}

	// Reset the bytes length to zero. This keeps the capacity but allows the slice to be reused.
	// Security note: This does not clear any data previously stored in the slice.
	sizedBytes := (*gotBytes)[:0]

	if cap(sizedBytes) < desiredCapacity {
		// We do not have a slice large enough, so we need to put the slice back into the pool
		// and then create a new byte slice with the desired size.
		Put(sizedBytes)
		sizedBytes = make([]byte, 0, desiredCapacity)
	}

	return sizedBytes
}

func Put(bytes []byte) {
	if bytes != nil {
		poolMutex.Lock()
		poolBytes.Put(&bytes)
		poolMutex.Unlock()
	}
}
