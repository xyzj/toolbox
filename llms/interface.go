package llms

import (
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
	Export(map[string]*ChatData) error
	Import() (map[string]*ChatData, error)
	Update(*ChatData) error
}
