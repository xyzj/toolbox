package storage

import "github.com/xyzj/toolbox/llms"

type MemStorage struct{}

func NewMemStorage() llms.Storage {
	return &MemStorage{}
}

func (s *MemStorage) Init() error {
	return nil
}

func (s *MemStorage) Export(m map[string]*llms.ChatData) error {
	return nil
}

func (s *MemStorage) Import() (map[string]*llms.ChatData, error) {
	return make(map[string]*llms.ChatData), nil
}

func (s *MemStorage) Update(*llms.ChatData) error {
	return nil
}
