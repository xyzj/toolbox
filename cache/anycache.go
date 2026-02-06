package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

type cacheData[T any] struct {
	locker sync.RWMutex
	data   map[string]*cData[T]
}

func (cd *cacheData[T]) len() int {
	cd.locker.RLock()
	defer cd.locker.RUnlock()
	return len(cd.data)
}
func (cd *cacheData[T]) clear() {
	cd.locker.Lock()
	defer cd.locker.Unlock()
	cd.data = make(map[string]*cData[T])
}
func (cd *cacheData[T]) store(key string, value T, expire time.Time) {
	cd.locker.Lock()
	defer cd.locker.Unlock()
	cd.data[key] = &cData[T]{data: value, expire: expire}
}
func (cd *cacheData[T]) load(key string) (T, bool) {
	cd.locker.RLock()
	defer cd.locker.RUnlock()
	v, ok := cd.data[key]
	if !ok {
		var x T
		return x, false
	}
	if v.expire.Before(time.Now()) {
		var x T
		return x, false
	}
	return v.data, true
}
func (cd *cacheData[T]) delete(key ...string) {
	cd.locker.Lock()
	defer cd.locker.Unlock()
	for _, k := range key {
		delete(cd.data, k)
	}
}
func (cd *cacheData[T]) clone() map[string]*cData[T] {
	cd.locker.RLock()
	defer cd.locker.RUnlock()
	cloneData := make(map[string]*cData[T], len(cd.data))
	for k, v := range cd.data {
		cloneData[k] = v
	}
	return cloneData
}
func (cd *cacheData[T]) expire(key string, expire time.Time) bool {
	cd.locker.Lock()
	defer cd.locker.Unlock()
	if v, ok := cd.data[key]; ok {
		v.expire = expire
		return true
	}
	return false
}

func (cd *cacheData[T]) foreach(f func(key string, value T) bool) {
	cd.locker.RLock()
	defer cd.locker.RUnlock()
	for k, v := range cd.data {
		if time.Now().After(v.expire) {
			continue
		}
		if !f(k, v.data) {
			break
		}
	}
}
func (cd *cacheData[T]) clearExpired() map[string]T {
	cd.locker.Lock()
	defer cd.locker.Unlock()
	now := time.Now()
	expired := make(map[string]T, len(cd.data))
	for k, v := range cd.data {
		if now.After(v.expire) {
			expired[k] = v.data
			delete(cd.data, k)
		}
	}
	return expired
}

type cData[T any] struct {
	expire time.Time
	data   T
}

// AnyCache 泛型结构缓存
type AnyCache[T any] struct {
	cache           *cacheData[T]
	cacheCleanup    *time.Ticker
	cleanupInterval time.Duration
	cacheExpire     time.Duration
	closed          bool
	closeCtx        context.Context
	closeFunc       context.CancelFunc
	closeOnce       sync.Once
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
	ctx, cancel := context.WithCancel(context.Background())
	x := &AnyCache[T]{
		cacheExpire:  expire,
		cache:        &cacheData[T]{data: make(map[string]*cData[T])},
		cacheCleanup: time.NewTicker(time.Minute),
		closeCtx:     ctx,
		closeFunc:    cancel,
		closeOnce:    sync.Once{},
	}
	go loopfunc.LoopFunc(func(params ...any) {
		for {
			select {
			case <-x.closeCtx.Done():
				return
			case <-x.cacheCleanup.C:
				expired := x.cache.clearExpired()
				if len(expired) > 0 && expireFunc != nil {
					loopfunc.GoFunc(func(params ...any) {
						expireFunc(expired)
					}, "expire func", logger.NewConsoleWriter())
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
	ac.closeOnce.Do(func() {
		ac.closed = true
		ac.cacheCleanup.Stop()
		ac.closeFunc()
		ac.cache.clear()
	})
}

// Clear clears all the entries from the cache.
// If the cache is already closed, this function does nothing.
//
// This function is safe to call concurrently with other methods of the AnyCache.
func (ac *AnyCache[T]) Clear() {
	if ac.closed {
		return
	}
	ac.cache.clear()
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
	if ac.closed {
		return 0
	}
	return ac.cache.len()
}

// Extension extends the expiration time of the specified cache entry by the cache's default expiration duration.
// If the cache entry does not exist, this function does nothing.
//
// Parameters:
//   - key: The key of the cache entry to be extended.
//
// This function is safe to call concurrently with other methods of the AnyCache.
func (ac *AnyCache[T]) Extension(key string) {
	ac.cache.expire(key, time.Now().Add(ac.cacheExpire))
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
	if ac.closed {
		return fmt.Errorf("cache is closed")
	}
	if !ac.cache.expire(key, time.Now().Add(expire)) {
		ac.cache.store(key, value, time.Now().Add(expire))
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
	if ac.closed {
		var zero T
		return zero, false
	}
	v, ok := ac.cache.load(key)
	if !ok {
		var zero T
		return zero, false
	}
	return v, true
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
	if ac.closed {
		var zero T
		return zero, false
	}
	v, ok := ac.cache.load(key)
	if !ok {
		ac.cache.store(key, value, time.Now().Add(ac.cacheExpire))
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
	if ac.closed {
		return
	}
	ac.cache.delete(key)
}

func (ac *AnyCache[T]) DeleteMore(keys ...string) {
	if ac.closed {
		return
	}
	ac.cache.delete(keys...)
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
	ac.cache.foreach(func(key string, value T) bool {
		return f(key, value)
	})
}
