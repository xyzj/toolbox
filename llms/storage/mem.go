package storage

import (
	"time"

	"github.com/xyzj/toolbox/llms"
)

type MemStorage struct{}

func NewMemStorage() llms.Storage {
	return &MemStorage{}
}

func (s *MemStorage) Clear() {}

func (s *MemStorage) Store(chat *llms.ChatData) error { return nil }

func (s *MemStorage) Load() (map[string]*llms.ChatData, error) {
	return make(map[string]*llms.ChatData), nil
}
func (s *MemStorage) RemoveDead(t time.Duration) {}
