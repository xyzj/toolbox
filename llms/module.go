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

type ChatResponse struct {
	Message    *Message `json:"message,omitempty"`
	Model      string   `json:"model"`
	CreatedAt  string   `json:"created_at"`
	Response   string   `json:"response,omitempty"` // 非流式响应时，回复内容在此字段
	DoneReason string   `json:"done_reason"`
	Done       bool     `json:"done"`
}

type ChatData struct {
	Messages   []*Message `json:"history"`
	ServerAddr string     `json:"uri"`
	Model      string     `json:"model"`
	ID         string     `json:"id"`
	MaxContext int        `json:"max_context"`
	ChatType   ChatType   `json:"chat_type"`
}

func (cd *ChatData) Marshal() []byte {
	s, _ := json.Marshal(cd)
	return s
}

type ChatType byte

const (
	Ollama ChatType = iota
)
