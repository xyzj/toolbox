package llms

import (
	"time"
)

type Chat interface {
	ID() string
	Chat(string, func([]byte) error) error
	Stop()
	Print() *ChatData
	Restore(*ChatData)
}

type Storage interface {
	Init() error
	RemoveDead(time.Duration)
	Import() (map[string]*ChatData, error)
	Update(*ChatData) error
}
