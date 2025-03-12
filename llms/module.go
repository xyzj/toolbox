package llms

import (
	"github.com/xyzj/toolbox/json"
)

type ChatRequest struct {
	Messages []*Message `json:"messages"`
	Model    string     `json:"model"`
	Stream   bool       `json:"stream"` // 设置为 false 获取非流式响应
}

func (cr *ChatRequest) Marshal() []byte {
	s, _ := json.Marshal(cr)
	return s
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatData struct {
	Messages   []*Message `json:"history"`
	ServerAddr string     `json:"uri"`
	Model      string     `json:"model"`
	ID         string     `json:"id"`
	ApiKey     string     `json:"api_key"`
	Timeout    int64      `json:"timeout"`
	LastUpdate int64      `json:"last_update"`
	MaxContext int        `json:"max_context"`
	ChatType   ChatType   `json:"chat_type"`
}

func (cd *ChatData) ToJSON() string {
	s, _ := json.MarshalToString(cd)
	return s
}

func (cd *ChatData) FromJSON(s string) error {
	return json.UnmarshalFromString(s, cd)
}

type ChatType byte

const (
	Ollama ChatType = iota
	VolcEngine
)
