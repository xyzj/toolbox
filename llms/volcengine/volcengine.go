package volcengine

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

var (
	idHash  = crypto.NewHash(crypto.HashSHA1, []byte{})
	ssdDone = "[DONE]"
)

type ChatResponse struct {
	Choices []*StreamChoice `json:"choices"`
	ID      string          `json:"id"`
	Model   string          `json:"model"`
	Created int64           `json:"created"`
}
type StreamChoice struct {
	Delta        *llms.Message `json:"delta,omitempty"`
	FinishReason string        `json:"finish_reason"` // stop,length,content_filter,tool_calls
	Index        int           `json:"index"`
}
type Opt struct {
	serverAddr string
	model      string
	apiKey     string
	timeout    time.Duration
	maxContent int
}
type Opts func(opt *Opt)

func OptServerAddr(s string) Opts {
	return func(o *Opt) {
		s = strings.TrimSuffix(s, "/")
		if !strings.HasSuffix(s, "/api/v3/chat/completions") {
			s += "/api/v3/chat/completions"
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

func OptAPIKey(s string) Opts {
	return func(o *Opt) {
		o.apiKey = s
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.opt.apiKey)
	c.buf.Reset()
	var err error
	r := &ChatResponse{}
	return c.client.DoStreamRequest(req, nil, func(b []byte) error {
		if c.stop.Load() {
			return errors.New("stop reading chat response")
		}
		b = bytes.TrimSpace(bytes.TrimPrefix(b, []byte("data: ")))
		if len(b) < 2 {
			return nil
		}
		if json.String(b) == ssdDone { // 结束
			if c.buf.Len() > 0 {
				c.history.Store(&llms.Message{
					Role:    "assistant",
					Content: c.buf.String(),
				})
			}
			return nil
		}
		err = json.Unmarshal(b, r)
		if err != nil {
			return errors.New("response data Unmarshal error:" + err.Error())
		}
		for _, chat := range r.Choices {
			if chat.Delta.Role != "assistant" {
				continue
			}
			if chat.FinishReason == "" {
				c.buf.WriteString(chat.Delta.Content)
				if f != nil {
					if err = f(json.Bytes(chat.Delta.Content)); err != nil {
						break
					}
				}
			}
		}
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
		apiKey:     d.ApiKey,
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
		ChatType:   llms.VolcEngine,
		ID:         c.id,
		ApiKey:     c.opt.apiKey,
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
		serverAddr: "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
		model:      "doubao-1-5-lite-32k-250115",
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
