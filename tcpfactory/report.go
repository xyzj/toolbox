package tcpfactory

import (
	"sync"
	"time"
)

type reportItem struct {
	id       uint64
	dtReport time.Time
	msg      any
	status   bool
}

// reportData 泛型结构缓存
type reportData struct {
	locker sync.RWMutex
	cache  map[uint64]*reportItem
}

// Store saves an item to the cache
func (rc *reportData) Store(key uint64, value *reportItem) {
	rc.locker.Lock()
	rc.cache[key] = value
	rc.locker.Unlock()
}

// Load retrieves an item from the cache
func (rc *reportData) Load(key uint64) (*reportItem, bool) {
	rc.locker.RLock()
	defer rc.locker.RUnlock()
	v, ok := rc.cache[key]
	return v, ok
}

// Delete removes an item from the cache
func (rc *reportData) Delete(keys ...uint64) {
	if len(keys) == 0 {
		return
	}
	rc.locker.Lock()
	for _, key := range keys {
		delete(rc.cache, key)
	}
	rc.locker.Unlock()
}

// ForEach calls the provided function for each item in the cache
func (rc *reportData) ForEach(f func(key uint64, value *reportItem) bool) {
	rc.locker.RLock()
	defer rc.locker.RUnlock()
	for k, v := range rc.cache {
		if !f(k, v) {
			break
		}
	}
}

// Clear removes all items from the cache
func (rc *reportData) Clear() {
	rc.locker.Lock()
	rc.cache = make(map[uint64]*reportItem)
	rc.locker.Unlock()
}

// newReportData creates a new reportData instance
func newReportData(l int) *reportData {
	return &reportData{
		cache:  make(map[uint64]*reportItem, l),
		locker: sync.RWMutex{},
	}
}
