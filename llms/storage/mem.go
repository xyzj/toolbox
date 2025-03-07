package storage

import (
	"time"

	"github.com/xyzj/toolbox/llms"
)

type MemStorage struct{}

func NewMemStorage() llms.Storage {
	return &MemStorage{}
}

func (s *MemStorage) Init() error {
	return nil
}

func (s *MemStorage) Clear(time.Duration) {
	return
}

func (s *MemStorage) Import() (map[string]*llms.ChatData, error) {
	return make(map[string]*llms.ChatData), nil
}

func (s *MemStorage) Update(*llms.ChatData) error {
	return nil
}
