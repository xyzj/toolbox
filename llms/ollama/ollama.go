package ollama

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/httpclient"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/llms"
	"github.com/xyzj/toolbox/llms/history"
)

var idHash = crypto.NewHash(crypto.HashSHA1, []byte{})

type Opt struct {
	ServerAddr string `json:"uri"`
	Model      string `json:"model"`
	MaxContent int
}

type Chat struct {
	id      string
	locker  sync.Mutex
	stop    atomic.Bool
	client  *httpclient.Client
	opt     *Opt
	history *history.ChatHistory
	buf     *bytes.Buffer
}

func (c *Chat) ID() string {
	return c.id
}

func (c *Chat) Chat(message string, f func([]byte) error, opts ...httpclient.ReqOpts) error {
	c.stop.Store(false)
	c.locker.Lock()
	defer c.locker.Unlock()
	// 生成用户问题，加入历史
	c.history.Store(&llms.Message{
		Role:    "user",
		Content: message,
	})

	// 组装聊天数据
	data := &llms.ChatRequest{
		Messages: c.history.Slice(),
		Model:    c.opt.Model,
		Stream:   true, // 流式响应，设置为 false 获取非流式响应
	}
	req, _ := http.NewRequest("POST", c.opt.ServerAddr, bytes.NewReader(data.Marshal()))
	buf := &bytes.Buffer{}
	var err error
	r := &llms.ChatResponse{}
	return c.client.DoStreamRequest(req, nil, func(b []byte) error {
		if c.stop.Load() {
			return errors.New("stop reading chat response")
		}
		err = json.Unmarshal(b, r)
		if err != nil {
			return errors.New("Ollama data Unmarshal error:" + err.Error())
		}
		if r.Message.Role != "assistant" {
			return nil
		}
		if !r.Done {
			buf.WriteString(r.Message.Content)
			if f != nil {
				return f(json.Bytes(r.Message.Content))
			}
			return nil
		}
		c.history.Store(&llms.Message{
			Role:    "assistant",
			Content: buf.String(),
		})
		return nil
	}, opts...)
}

func (c *Chat) Stop() {
	c.stop.Store(true)
}

func (c *Chat) SetID(id string) {
	c.id = id
}

func (c *Chat) Restore(d *llms.ChatData) {
	c.id = d.ID
	c.opt = &Opt{
		ServerAddr: d.ServerAddr,
		Model:      d.Model,
		MaxContent: d.MaxContext,
	}
	if c.history == nil {
		c.history = history.NewChatHistory(d.MaxContext)
	}
	c.history.StoreMany(d.Messages...)
	c.client = httpclient.New()
	c.locker = sync.Mutex{}
	c.stop = atomic.Bool{}
	c.buf = &bytes.Buffer{}
}

func (c *Chat) Print() *llms.ChatData {
	return &llms.ChatData{
		ChatType:   llms.Ollama,
		ID:         c.id,
		Messages:   c.history.Slice(),
		Model:      c.opt.Model,
		ServerAddr: c.opt.ServerAddr,
		MaxContext: c.opt.MaxContent,
		LastUpdate: time.Now().Unix(),
	}
}

func NewChat(opt *Opt) *Chat {
	if opt == nil {
		opt = &Opt{}
	}
	if opt.ServerAddr == "" {
		opt.ServerAddr = "http://localhost:11434"
	}
	opt.ServerAddr = strings.TrimSuffix(opt.ServerAddr, "/") + "/api/chat"
	if opt.Model == "" {
		opt.Model = "default"
	}
	opt.MaxContent = min(max(opt.MaxContent, 100), 1000)
	return &Chat{
		id:      idHash.Hash([]byte(time.Now().String())).HexString(),
		client:  httpclient.New(),
		locker:  sync.Mutex{},
		stop:    atomic.Bool{},
		buf:     &bytes.Buffer{},
		opt:     opt,
		history: history.NewChatHistory(opt.MaxContent),
	}
}
