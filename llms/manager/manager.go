package manager

import (
	"errors"
	"time"

	"github.com/xyzj/toolbox/cache"
	"github.com/xyzj/toolbox/httpclient"
	"github.com/xyzj/toolbox/llms"
	"github.com/xyzj/toolbox/llms/ollama"
	"github.com/xyzj/toolbox/llms/storage"
	"github.com/xyzj/toolbox/llms/volcengine"
	"github.com/xyzj/toolbox/logger"
)

type (
	Opt struct {
		dataStorage  llms.Storage
		chatLifeTime time.Duration
		logg         logger.Logger
	}
	Opts func(opt *Opt)

	ChatManager struct {
		data  llms.Storage
		chats *cache.AnyCache[llms.Chat]
		opt   *Opt
	}
)

func OptStorage(s llms.Storage) Opts {
	return func(o *Opt) {
		o.dataStorage = s
	}
}

func OptChatLifeTime(t time.Duration) Opts {
	return func(o *Opt) {
		o.chatLifeTime = t
	}
}

func OptLogger(l logger.Logger) Opts {
	return func(o *Opt) {
		o.logg = l
	}
}

func (m *ChatManager) Store(chat llms.Chat) {
	m.chats.Store(chat.ID(), chat)
}

func (m *ChatManager) Chat(id, message string, f func([]byte) error, opts ...httpclient.ReqOptions) error {
	chat, ok := m.chats.Load(id)
	if !ok {
		return errors.New("chat not found")
	}
	err := chat.Chat(message, f)
	if err != nil {
		m.opt.logg.Error("Request chat error: " + err.Error())
		return errors.New("Request chat error: " + err.Error())
	}
	m.chats.Store(chat.ID(), chat)
	// save the chat
	err = m.data.Store(chat.Print())
	if err != nil {
		m.opt.logg.Error("Update storage error: " + err.Error())
		return errors.New("Update storage error: " + err.Error())
	}
	return nil
}

func (m *ChatManager) ChatRaw(id, message string, f func([]byte) error, opts ...httpclient.ReqOptions) error {
	chat, ok := m.chats.Load(id)
	if !ok {
		return errors.New("chat not found")
	}
	err := chat.ChatRaw(message, f)
	if err != nil {
		m.opt.logg.Error("Request chat error: " + err.Error())
		return errors.New("Request chat error: " + err.Error())
	}
	m.chats.Store(chat.ID(), chat)
	// save the chat
	err = m.data.Store(chat.Print())
	if err != nil {
		m.opt.logg.Error("Update storage error: " + err.Error())
		return errors.New("Update storage error: " + err.Error())
	}
	return nil
}

func (m *ChatManager) Stop(id string) error {
	chat, ok := m.chats.Load(id)
	if !ok {
		return errors.New("chat not found")
	}
	chat.Stop()
	return nil
}

func (m *ChatManager) Load() {
	chats, err := m.data.Load()
	if err != nil {
		m.opt.logg.Error("Load storage error: " + err.Error())
		return
	}
	for id, chat := range chats {
		if time.Since(time.Unix(chat.LastUpdate, 0)) > m.opt.chatLifeTime {
			continue
		}
		switch chat.ChatType {
		case llms.Ollama:
			c := &ollama.Chat{}
			c.Restore(chat)
			m.chats.Store(id, c)
		case llms.VolcEngine:
			c := &volcengine.Chat{}
			c.Restore(chat)
			m.chats.Store(id, c)
		}
	}
}

func NewChatManager(opts ...Opts) *ChatManager {
	opt := &Opt{
		chatLifeTime: time.Hour * 24,
		logg:         logger.NewNilLogger(),
		dataStorage:  storage.NewMemStorage(),
	}
	for _, o := range opts {
		o(opt)
	}
	c := &ChatManager{
		opt:   opt,
		chats: cache.NewAnyCache[llms.Chat](opt.chatLifeTime),
		data:  opt.dataStorage,
	}
	go func() {
		t := time.NewTicker(time.Minute * 5)
		for range t.C {
			c.data.RemoveDead(opt.chatLifeTime)
		}
	}()
	c.Load()
	return c
}
