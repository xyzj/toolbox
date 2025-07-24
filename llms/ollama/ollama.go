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

var idHash = crypto.NewHash(crypto.HashSHA1)

type ChatResponse struct {
	Message    *llms.Message `json:"message,omitempty"`
	Model      string        `json:"model"`
	CreatedAt  string        `json:"created_at"`
	Response   string        `json:"response,omitempty"` // 非流式响应时，回复内容在此字段
	DoneReason string        `json:"done_reason"`
	Done       bool          `json:"done"`
}
type Opt struct {
	serverAddr string
	model      string
	timeout    time.Duration
	maxContent int
}

type Opts func(opt *Opt)

func OptServerAddr(s string) Opts {
	return func(o *Opt) {
		s = strings.TrimSuffix(s, "/")
		if !strings.HasSuffix(s, "/api/chat") {
			s += "/api/chat"
		}
		o.serverAddr = s
	}
}

func OptModel(s string) Opts {
	return func(o *Opt) {
		if s != "" {
			o.model = s
		}
	}
}

func OptMaxContent(n int) Opts {
	return func(o *Opt) {
		o.maxContent = min(max(n, 100), 1000)
	}
}

func OptTimeout(t time.Duration) Opts {
	return func(o *Opt) {
		o.timeout = t
	}
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

func (c *Chat) Chat(message string, f func([]byte) error) error {
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
		Model:    c.opt.model,
		Stream:   true, // 流式响应，设置为 false 获取非流式响应
	}
	req, _ := http.NewRequest("POST", c.opt.serverAddr, bytes.NewReader(data.Marshal()))
	c.buf.Reset()
	var err error
	r := &ChatResponse{}
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
			c.buf.WriteString(r.Message.Content)
			if f != nil {
				return f(json.Bytes(r.Message.Content))
			}
			return nil
		}
		c.history.Store(&llms.Message{
			Role:    "assistant",
			Content: c.buf.String(),
		})
		return nil
	}, httpclient.OptTimeout(c.opt.timeout))
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
		serverAddr: d.ServerAddr,
		model:      d.Model,
		maxContent: d.MaxContext,
		timeout:    time.Second * time.Duration(d.Timeout),
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
		Model:      c.opt.model,
		ServerAddr: c.opt.serverAddr,
		MaxContext: c.opt.maxContent,
		LastUpdate: time.Now().Unix(),
		Timeout:    int64(c.opt.timeout.Seconds()),
	}
}

func NewChat(opts ...Opts) *Chat {
	opt := &Opt{
		serverAddr: "http://localhost:11434",
		model:      "gemma2",
		maxContent: 1000,
		timeout:    time.Minute * 3,
	}
	for _, o := range opts {
		o(opt)
	}
	return &Chat{
		id:      idHash.Hash([]byte(time.Now().String())).HexString(),
		client:  httpclient.New(),
		locker:  sync.Mutex{},
		stop:    atomic.Bool{},
		buf:     &bytes.Buffer{},
		opt:     opt,
		history: history.NewChatHistory(opt.maxContent),
	}
}
