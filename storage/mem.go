package storage

import "sync"

type Memory struct {
	mu       sync.RWMutex
	maxSize  int
	historys []string
}

func NewMemory(maxSize int) *Memory {
	if maxSize < 0 {
		maxSize = 0
	}

	return &Memory{maxSize: maxSize}
}

func (m *Memory) Store(history ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(history) == 0 {
		return nil
	}

	if m.maxSize == 0 {
		m.historys = nil
		return nil
	}

	m.historys = append(m.historys, history...)
	if len(m.historys) > m.maxSize {
		overSize := len(m.historys) - m.maxSize
		copy(m.historys, m.historys[overSize:])
		m.historys = m.historys[:m.maxSize]
	}

	return nil
}

func (m *Memory) Load() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]string(nil), m.historys...), nil
}

func (m *Memory) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.historys = nil
	return nil
}
