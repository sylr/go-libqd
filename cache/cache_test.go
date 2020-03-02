package cache

import (
	"sync"
	"testing"
	"time"
)

func TestNewCacher(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(100 * 100)

	for i := int64(1); i <= 100; i++ {
		duration := time.Duration(i) * time.Minute
		for j := int64(1); j <= 100; j++ {
			cleanup := time.Duration(j) * time.Minute

			go func(duration time.Duration, cleanup time.Duration) {
				GetCache(duration, cleanup)
				wg.Done()
			}(duration, cleanup)
		}
	}

	wg.Wait()
}

func BenchmarkNewCacher(b *testing.B) {
	wg := sync.WaitGroup{}
	wg.Add(100 * 100)

	for i := int64(1); i <= 100; i++ {
		duration := time.Duration(i) * time.Minute
		for j := int64(1); j <= 100; j++ {
			cleanup := time.Duration(j) * time.Minute

			go func(duration time.Duration, cleanup time.Duration) {
				GetCache(duration, cleanup)
				wg.Done()
			}(duration, cleanup)
		}
	}

	wg.Wait()
}
