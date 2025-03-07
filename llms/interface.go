package llms

import (
	"time"

	"github.com/xyzj/toolbox/httpclient"
)

type Chat interface {
	ID() string
	Chat(string, func([]byte) error, ...httpclient.ReqOpts) error
	Stop()
	Print() *ChatData
	Restore(*ChatData)
}

type Storage interface {
	Init() error
	Clear(time.Duration)
	Import() (map[string]*ChatData, error)
	Update(*ChatData) error
}
