package batcher_test

import (
	"sync"
	"testing"

	"diskey/pkg/batcher"
)

func TestRun_concurrent(t *testing.T) {
	t.Parallel()

	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			waitGroup.Done()
		}
	})

	for i := 0; i < batchSize*1000; i++ {
		waitGroup.Add(1)
		batchChannel <- i
	}

	close(batchChannel)
	waitGroup.Wait()

	t.Fail()
}

func TestRun_sequential(t *testing.T) {
	t.Parallel()

	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			waitGroup.Done()
		}
	})

	for i := 0; i < batchSize*1000; i++ {
		waitGroup.Add(1)
		batchChannel <- i
		waitGroup.Wait()
	}

	close(batchChannel)
	waitGroup.Wait()

	t.Fail()
}
