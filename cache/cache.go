package cache

import (
	"sync"
	"time"

	cache "github.com/sylr/go-cache/v2"
)

var (
	noopCaching = false
	noopCacher  cache.Cacher
)

var (
	mutex  = sync.RWMutex{}
	caches = make(map[time.Duration]map[time.Duration]cache.Cacher)
)

// SetNoop tells wheter or not next call of GetCache would return a NoopCacher
func SetNoop(noop bool) {
	mutex.Lock()
	defer mutex.Unlock()

	noopCaching = noop
}

// GetCache returns a caching object
func GetCache(duration time.Duration, cleanupInterval time.Duration) cache.Cacher {
	mutex.Lock()
	defer mutex.Unlock()

	if noopCaching {
		if noopCacher == nil {
			noopCacher = cache.NewNoopCacher(time.Minute, time.Minute)
		}

		return noopCacher
	}

	if _, ok := caches[duration]; !ok {
		caches[duration] = make(map[time.Duration]cache.Cacher)
	}

	if _, ok := caches[duration][cleanupInterval]; !ok {
		caches[duration][cleanupInterval] = cache.NewCacher(duration, cleanupInterval)
	}

	return caches[duration][cleanupInterval]
}
