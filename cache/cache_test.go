package cache

import (
	"sync"
	"testing"
	"time"

	cache "github.com/sylr/go-cache/v2"
)

func TestGetCache(t *testing.T) {
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

// -- benchmarks ---------------------------------------------------------------

type getCacheFunc func(time.Duration, time.Duration) cache.Cacher

var (
	getCacheFuncs = []struct {
		name string
		fun  getCacheFunc
	}{
		{"Cache", GetCache},
		{"MeteredCache", GetMeteredCache},
	}
)

func benchGetCache(b *testing.B, f getCacheFunc, durations int, cleanups int) {
	wg := sync.WaitGroup{}
	wg.Add(durations * cleanups)

	b.StartTimer()
	for i := int64(1); i <= int64(durations); i++ {
		duration := time.Duration(i) * time.Minute
		for j := int64(1); j <= int64(cleanups); j++ {
			cleanup := time.Duration(j) * time.Minute

			go func(duration time.Duration, cleanup time.Duration) {
				f(duration, cleanup)
				wg.Done()
			}(duration, cleanup)
		}
	}

	wg.Wait()
	b.StopTimer()
}

func BenchmarkGetCache(b *testing.B) {
	b.StopTimer()
	for _, f := range getCacheFuncs {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resetCaches()
				benchGetCache(b, f.fun, 10, 20)
			}
		})
	}
}

func benchGetCacheAdd(b *testing.B, f getCacheFunc, rounds int) {
	c := f(time.Duration(b.N)*time.Minute, time.Duration(b.N)*time.Minute)

	b.StartTimer()
	for i := 0; i < rounds; i++ {
		c.Add(string(byte(i%100)), i, time.Minute)
	}
	b.StopTimer()
}

func BenchmarkGetCacheAdd(b *testing.B) {
	b.StopTimer()
	for _, f := range getCacheFuncs {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resetCaches()
				benchGetCacheAdd(b, f.fun, 200)
			}
		})
	}
}

func benchGetCacheIncrement(b *testing.B, f getCacheFunc, rounds int) {
	wg := sync.WaitGroup{}
	c := f(time.Duration(b.N)*time.Minute, time.Duration(b.N)*time.Minute)

	for i := 0; i < 10; i++ {
		c.Add(string(byte(i%10)), i, time.Minute)
	}

	wg.Add(rounds)

	b.StartTimer()
	for i := 0; i < rounds; i++ {
		go func(i int) {
			c.IncrementInt(string(byte(i%10)), i)
			wg.Done()
		}(i)
	}

	wg.Wait()
	b.StopTimer()
}

func BenchmarkGetCacheIncrement(b *testing.B) {
	b.StopTimer()
	for _, f := range getCacheFuncs {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resetCaches()
				benchGetCacheIncrement(b, f.fun, 200)
			}
		})
	}
	b.StopTimer()
}

func benchGetCacheDecrement(b *testing.B, f getCacheFunc, rounds int) {
	wg := sync.WaitGroup{}
	c := f(time.Duration(b.N)*time.Minute, time.Duration(b.N)*time.Minute)

	for i := 0; i < 10; i++ {
		c.Add(string(byte(i%10)), i, time.Minute)
	}

	wg.Add(rounds)

	b.StartTimer()
	for i := 0; i < rounds; i++ {
		go func(i int) {
			c.DecrementInt(string(byte(i%10)), i)
			wg.Done()
		}(i)
	}

	wg.Wait()
	b.StopTimer()
}

func BenchmarkGetCacheDecrement(b *testing.B) {
	b.StopTimer()
	for _, f := range getCacheFuncs {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resetCaches()
				benchGetCacheDecrement(b, f.fun, 200)
			}
		})
	}
	b.StopTimer()
}
