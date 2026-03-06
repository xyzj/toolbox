package storage

import (
	"sync"
)

type Ring struct {
	mu      sync.RWMutex
	maxSize int
	start   int
	size    int
	buf     []string
}

func NewRing(maxSize int) *Ring {
	if maxSize < 0 {
		maxSize = 0
	}

	return &Ring{maxSize: maxSize}
}

func (r *Ring) Store(history ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(history) == 0 {
		return nil
	}

	if r.maxSize == 0 {
		r.buf = nil
		r.start = 0
		r.size = 0
		return nil
	}

	if len(r.buf) == 0 {
		r.buf = make([]string, r.maxSize)
	}

	for _, item := range history {
		if r.size < r.maxSize {
			idx := (r.start + r.size) % r.maxSize
			r.buf[idx] = item
			r.size++
			continue
		}

		r.buf[r.start] = item
		r.start = (r.start + 1) % r.maxSize
	}

	return nil
}

func (r *Ring) Load() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.size == 0 || len(r.buf) == 0 {
		return nil, nil
	}

	historys := make([]string, r.size)
	firstPart := r.maxSize - r.start
	if firstPart > r.size {
		firstPart = r.size
	}

	copy(historys, r.buf[r.start:r.start+firstPart])
	if firstPart < r.size {
		copy(historys[firstPart:], r.buf[:r.size-firstPart])
	}

	return historys, nil
}

func (r *Ring) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.size; i++ {
		idx := (r.start + i) % r.maxSize
		r.buf[idx] = ""
	}
	r.start = 0
	r.size = 0
	return nil
}
