package cache

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/mapfx"
)

var defautlCacheCleanup = time.Second * 60

// SetDefaultCacheCleanupTime sets the default time interval for cache cleanup.
// The cleanup interval is used by the AnyCache to periodically check for expired entries.
//
// Parameters:
// - t: The time duration for cache cleanup.
//
// This function does not return any value.
func SetDefaultCacheCleanupTime(t time.Duration) {
	defautlCacheCleanup = t
}

type cData[T any] struct {
	expire time.Time
	data   T
}

// AnyCache 泛型结构缓存
type AnyCache[T any] struct {
	cache        *mapfx.StructMap[string, cData[T]]
	cacheCleanup *time.Ticker
	cacheExpire  time.Duration
	closed       atomic.Bool
	closeChan    chan bool
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
		cacheCleanup: time.NewTicker(defautlCacheCleanup),
		closeChan:    make(chan bool, 1),
	}
	x.closed.Store(false)
	go loopfunc.LoopFunc(func(params ...interface{}) {
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
						loopfunc.GoFunc(func(params ...interface{}) {
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

// SetCleanUp 设置清理周期，不低于1秒
func (ac *AnyCache[T]) SetCleanUp(cleanup time.Duration) {
	if cleanup < time.Second {
		cleanup = time.Second
	}
	ac.cacheCleanup.Reset(cleanup)
}

// Close 关闭这个缓存，如果需要再次使用，应调用NewAnyCache方法重新初始化
func (ac *AnyCache[T]) Close() {
	ac.closed.Store(true)
	ac.cacheCleanup.Stop()
	ac.closeChan <- true
	ac.cache.Clean()
	ac.cache = nil
}

// Clean 清空这个缓存
func (ac *AnyCache[T]) Clean() {
	if ac.closed.Load() {
		return
	}
	ac.cache.Clean()
}

// Len 返回缓存内容数量
func (ac *AnyCache[T]) Len() int {
	if ac.closed.Load() {
		return 0
	}
	return ac.cache.Len()
}

// Extension 将指定缓存延期
func (ac *AnyCache[T]) Extension(key string) {
	if x, ok := ac.cache.LoadForUpdate(key); ok {
		x.expire = time.Now().Add(ac.cacheExpire)
	}
}

// Store 添加缓存内容，如果缓存已关闭，会返回错误
func (ac *AnyCache[T]) Store(key string, value T) error {
	return ac.StoreWithExpire(key, value, ac.cacheExpire)
}

// StoreWithExpire 添加缓存内容，设置自定义的有效时间，如果缓存已关闭，会返回错误
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

// Load 读取一个缓存内容，如果不存在，返回false
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
		// ac.cache.Delete(key) // 删除会有锁操作，因此还是放在清理方法里一次性做
		return *x, false
	}
	return v.data, true
}

// LoadOrStore 读取或者设置一个缓存内如
//
//	当key存在时，返回缓存内容，并设置true
//	当key不存在时，将内容加入缓存，返回设置内容，并设置false
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

// Delete 删除一个缓存内容
func (ac *AnyCache[T]) Delete(key string) {
	if ac.closed.Load() {
		return
	}
	ac.cache.Delete(key)
}

// ForEach 遍历所有缓存内容
func (ac *AnyCache[T]) ForEach(f func(key string, value T) bool) {
	ac.cache.ForEach(func(key string, value *cData[T]) bool {
		if time.Now().After(value.expire) {
			return true
		}
		return f(key, value.data)
	})
}
