package batcher_test

import (
	"sync"
	"testing"
	"time"

	"diskey/pkg/batcher"
)

func BenchmarkRun_concurrent(b *testing.B) {
	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			time.Sleep(time.Nanosecond)
			waitGroup.Done()
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		waitGroup.Add(1)
		batchChannel <- i
	}

	close(batchChannel)
	waitGroup.Wait()

	b.StopTimer()
}

func BenchmarkRun_parallel(b *testing.B) {
	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			time.Sleep(time.Nanosecond)
			waitGroup.Done()
		}
	})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			waitGroup.Add(1)
			batchChannel <- 1
		}
	})

	close(batchChannel)
	waitGroup.Wait()

	b.StopTimer()
}

func BenchmarkRun_concurrent_slow_job(b *testing.B) {
	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			time.Sleep(100 * time.Millisecond)
			waitGroup.Done()
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		waitGroup.Add(1)
		batchChannel <- i
	}

	close(batchChannel)
	waitGroup.Wait()

	b.StopTimer()
}

func BenchmarkRun_parallel_slow_job(b *testing.B) {
	waitGroup := &sync.WaitGroup{}

	batchSize := 1000
	batchChannel := batcher.Run(batchSize, func(batch []int) {
		for range batch {
			time.Sleep(100 * time.Millisecond)
			waitGroup.Done()
		}
	})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			waitGroup.Add(1)
			batchChannel <- 1
		}
	})

	close(batchChannel)
	waitGroup.Wait()

	b.StopTimer()
}
