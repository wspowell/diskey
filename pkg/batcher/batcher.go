package batcher

import (
	"time"
)

func Run[T any](batchSize int, onFlush func(batch []T)) chan<- T {
	batchChannel := make(chan T, batchSize)

	const numberOfBuffers = 1
	for range numberOfBuffers {
		go processBatches(batchChannel, batchSize, onFlush)
	}

	return batchChannel
}

func processBatches[T any](batchChannel <-chan T, batchSize int, onFlush func(batch []T)) {
	maxFlushInterval := time.Second

	flushInterval := time.Millisecond

	var length int
	var item T
	batchedItems := make([]T, batchSize)
	ticker := time.NewTicker(flushInterval)
	running := true

	itemsSinceLastInterval := 0

	for {
		select {
		// A receive expression used in an assignment statement or initialization of the special form yields
		// an additional untyped boolean result reporting whether the communication succeeded. The value of
		// ok is true if the value received was delivered by a successful send operation to the channel, or
		// false if it is a zero value generated because the channel is closed and empty.
		case item, running = <-batchChannel:
			if !running {
				running = false
				ticker.Stop()
				// Break out of the select{} in order to trigger a flush and exit.
				// Do not add any more items to the queue because it will only be zero valued items.
				break
			}

			batchedItems[length] = item
			length++
			itemsSinceLastInterval++

			if length != batchSize {
				// Restart the loop and keep collection batch items.
				continue
			}
		case <-ticker.C:
			// Dynamically adjust the flush interval to reduce waits and contention.
			if itemsSinceLastInterval > 0 {
				batchesSinceLastInternal := float64(itemsSinceLastInterval) / float64(batchSize)
				// If the max interval is way to large, we can start from the fastest interval and scale up from there.
				newInterval := time.Nanosecond
				if batchesSinceLastInternal > 1 {
					intervalPerBatch := float64(flushInterval.Nanoseconds()) / batchesSinceLastInternal
					newInterval = time.Duration(int64(intervalPerBatch * float64(time.Nanosecond)))
				}

				if newInterval.Nanoseconds() > maxFlushInterval.Nanoseconds() {
					newInterval = maxFlushInterval
				}
				if newInterval.Nanoseconds() <= 0 {
					newInterval = time.Nanosecond
				}
				// fmt.Println("batches since last interval", batchesSinceLastInternal, "items since last interval", itemsSinceLastInterval, "updating interval", flushInterval.Nanoseconds(), "->", newInterval.Nanoseconds())
				flushInterval = newInterval

				itemsSinceLastInterval = 0

				// Break out of the select{} in order to trigger a flush.
				break
			}
		}

		if length != 0 {
			onFlush(batchedItems[:length])
			length = 0
		}

		ticker.Reset(flushInterval)

		if !running {
			// Stop the for{} and exit.
			break
		}
	}
}
