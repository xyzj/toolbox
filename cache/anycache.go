package cache

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/mapfx"
)

type cData[T any] struct {
	expire time.Time
	data   T
}

// AnyCache 泛型结构缓存
type AnyCache[T any] struct {
	cache           *mapfx.StructMap[string, cData[T]]
	cacheCleanup    *time.Ticker
	cleanupInterval time.Duration
	cacheExpire     time.Duration
	closed          atomic.Bool
	closeChan       chan bool
}

// NewAnyCacheWithExpireFunc initializes a new cache with a specified expiration time and an optional expiration function.
// The cache will create a goroutine to periodically check for expired entries.
// When the cache is no longer needed, it should be closed using the Close() method.
//
// Parameters:
//   - expire: The duration for which cache entries should be considered valid.
//   - expireFunc: An optional function to be executed when a cache entry expires.
//     The function will receive a map of expired entries, where the key is the entry key and the value is the entry data.
//
// Return:
// - A pointer to the newly created AnyCache instance.
func NewAnyCacheWithExpireFunc[T any](expire time.Duration, expireFunc func(map[string]T)) *AnyCache[T] {
	x := &AnyCache[T]{
		cacheExpire:  expire,
		cache:        mapfx.NewStructMap[string, cData[T]](),
		cacheCleanup: time.NewTicker(time.Minute),
		closeChan:    make(chan bool, 1),
	}
	x.closed.Store(false)
	go loopfunc.LoopFunc(func(params ...any) {
		for {
			select {
			case <-x.closeChan:
				return
			case <-x.cacheCleanup.C:
				tnow := time.Now()
				keys := make([]string, 0, x.cache.Len())
				ex := make(map[string]T)
				for k, v := range x.cache.Clone() {
					if tnow.After(v.expire) {
						keys = append(keys, k)
						ex[k] = v.data
					}
				}
				if len(keys) > 0 {
					x.cache.DeleteMore(keys...)
					if expireFunc != nil {
						loopfunc.GoFunc(func(params ...any) {
							expireFunc(ex)
						}, "expire func", logger.NewConsoleWriter())
					}
				}

			}
		}
	}, "any cache", logger.NewConsoleWriter())
	return x
}

// NewAnyCache initializes a new cache with a specified expiration time.
// The cache will create a goroutine to periodically check for expired entries.
// When the cache is no longer needed, it should be closed using the Close() method.
//
// Parameters:
//   - expire: The duration for which cache entries should be considered valid.
//
// Return:
// - A pointer to the newly created AnyCache instance.
//
// Example:
//
//	cache := NewAnyCache[int](time.Minute * 5)
//	defer cache.Close()
//	cache.Store("key1", 100)
//	value, ok := cache.Load("key1")
//	if ok {
//	    fmt.Println("Value:", value) // Output: Value: 100
//	}
func NewAnyCache[T any](expire time.Duration) *AnyCache[T] {
	return NewAnyCacheWithExpireFunc[T](expire, nil)
}

// SetCleanUp sets the cleanup period for the cache. The cleanup period should not be less than 1 second.
// If the cleanup period is less than 1 second, it will be automatically set to 1 second.
//
// Parameters:
// - cleanup: The duration for the cleanup period.
func (ac *AnyCache[T]) SetCleanUp(cleanup time.Duration) {
	if cleanup < time.Second {
		cleanup = time.Second
	}
	ac.cacheCleanup.Reset(cleanup)
}

// Close closes this cache. If the cache needs to be used again, it should be reinitialized using the NewAnyCache method.
// This method stops the cleanup goroutine, sends a signal to close the channel, clears the cache, and sets the cache pointer to nil.
func (ac *AnyCache[T]) Close() {
	ac.closed.Store(true)
	ac.cacheCleanup.Stop()
	ac.closeChan <- true
	ac.cache.Clear()
	ac.cache = nil
}

// Clear clears all the entries from the cache.
// If the cache is already closed, this function does nothing.
//
// This function is safe to call concurrently with other methods of the AnyCache.
func (ac *AnyCache[T]) Clear() {
	if ac.closed.Load() {
		return
	}
	ac.cache.Clear()
}

// Len returns the number of entries in the cache.
// If the cache is closed, it returns 0.
//
// This function is safe to call concurrently with other methods of the AnyCache.
//
// Parameters:
//   - ac: A pointer to the AnyCache instance.
//
// Return:
//   - An integer representing the number of entries in the cache.
func (ac *AnyCache[T]) Len() int {
	if ac.closed.Load() {
		return 0
	}
	return ac.cache.Len()
}

// Extension extends the expiration time of the specified cache entry by the cache's default expiration duration.
// If the cache entry does not exist, this function does nothing.
//
// Parameters:
//   - key: The key of the cache entry to be extended.
//
// This function is safe to call concurrently with other methods of the AnyCache.
func (ac *AnyCache[T]) Extension(key string) {
	if x, ok := ac.cache.LoadForUpdate(key); ok {
		x.expire = time.Now().Add(ac.cacheExpire)
	}
}

// Store adds a cache entry with the specified key and value.
// If the cache is already closed, it returns an error.
//
// Parameters:
//   - key: The unique identifier for the cache entry.
//   - value: The data to be stored in the cache.
//
// Return:
//   - An error if the cache is closed, otherwise nil.
func (ac *AnyCache[T]) Store(key string, value T) error {
	return ac.StoreWithExpire(key, value, ac.cacheExpire)
}

// StoreWithExpire adds a cache entry with the specified key, value, and expiration duration.
// If the cache is already closed, it returns an error.
//
// Parameters:
//   - key: The unique identifier for the cache entry.
//   - value: The data to be stored in the cache.
//   - expire: The duration for which the cache entry should be considered valid.
//
// Return:
//   - An error if the cache is closed, otherwise nil.
func (ac *AnyCache[T]) StoreWithExpire(key string, value T, expire time.Duration) error {
	if ac.closed.Load() {
		return fmt.Errorf("cache is closed")
	}
	if v, ok := ac.cache.LoadForUpdate(key); ok {
		v.expire = time.Now().Add(expire)
		v.data = value
	} else {
		ac.cache.Store(key, &cData[T]{
			expire: time.Now().Add(expire),
			data:   value,
		})
	}
	return nil
}

// Load retrieves the value associated with the given key from the cache.
// If the key is not found or the entry has expired, it returns the zero value of type T and false.
//
// Parameters:
//   - key: The unique identifier for the cache entry.
//
// Return:
//   - The value associated with the given key if found and not expired.
//   - A boolean value indicating whether the key was found and not expired.
func (ac *AnyCache[T]) Load(key string) (T, bool) {
	x := new(T)
	if ac.closed.Load() {
		return *x, false
	}
	v, ok := ac.cache.Load(key)
	if !ok {
		return *x, false
	}
	if time.Now().After(v.expire) {
		// ac.cache.Delete(key) // Deleting here would cause a lock operation, so it's done in the cleanup method instead.
		return *x, false
	}
	return v.data, true
}

// LoadOrStore reads or sets a cache entry.
//
// When the key exists:
// - Returns the cached content and sets the boolean return value to true.
//
// When the key does not exist:
// - Adds the content to the cache and returns the set content, along with the boolean return value set to false.
//
// Parameters:
// - key: The unique identifier for the cache entry.
// - value: The data to be stored in the cache.
//
// Return:
//   - The value associated with the given key if found and not expired.
//   - A boolean value indicating whether the key was found and not expired.
//     If the cache is closed, it returns the zero value of type T and false.
func (ac *AnyCache[T]) LoadOrStore(key string, value T) (T, bool) {
	x := new(T)
	if ac.closed.Load() {
		return *x, false
	}
	v, ok := ac.Load(key)
	if !ok {
		ac.cache.Store(key, &cData[T]{
			expire: time.Now().Add(ac.cacheExpire),
			data:   value,
		})
		return value, false
	}
	return v, true
}

// Delete removes a cache entry with the specified key.
// If the cache is already closed, this function does nothing.
//
// Parameters:
//   - key: The unique identifier for the cache entry to be deleted.
//
// Return:
//   - None
func (ac *AnyCache[T]) Delete(key string) {
	if ac.closed.Load() {
		return
	}
	ac.cache.Delete(key)
}

// ForEach iterates over all the entries in the cache and applies the provided function to each entry.
// The function will be called for each entry, excluding expired entries.
// If the function returns false, the iteration will be stopped.
//
// Parameters:
//   - f: A function that takes a key and a value as parameters and returns a boolean value.
//     The function will be called for each entry in the cache.
//
// Return:
//   - None
func (ac *AnyCache[T]) ForEach(f func(key string, value T) bool) {
	ac.cache.ForEach(func(key string, value *cData[T]) bool {
		if time.Now().After(value.expire) {
			return true
		}
		return f(key, value.data)
	})
}
