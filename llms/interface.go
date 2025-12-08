package llms

import "time"

type Chat interface {
	ID() string
	Chat(string, func([]byte) error) error
	ChatRaw(string, func([]byte) error) error
	Stop()
	Print() *ChatData
	Restore(*ChatData)
}

type Storage interface {
	Store(*ChatData) error
	Load() (map[string]*ChatData, error)
	Clear()
	RemoveDead(time.Duration)
}
