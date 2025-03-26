package history

import (
	"container/ring"

	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/llms"
)

func NewChatHistory(context int) *ChatHistory {
	return &ChatHistory{
		data:       ring.New(context),
		maxContext: context * 2,
	}
}

// ChatHistory 一个不重复的struct切片结构
type ChatHistory struct {
	// locker     sync.RWMutex
	data       *ring.Ring
	maxContext int
}

func (u *ChatHistory) Store(msg *llms.Message) bool {
	// u.locker.Lock()
	// defer u.locker.Unlock()
	u.data.Value = msg
	u.data.Next()
	return true
}

func (u *ChatHistory) StoreMany(msgs ...*llms.Message) {
	// u.locker.Lock()
	// defer u.locker.Unlock()
	for _, msg := range msgs {
		u.data.Value = msg
		u.data.Next()
	}
}

func (u *ChatHistory) Clear() {
	// u.locker.Lock()
	// defer u.locker.Unlock()
	u.data.Do(func(a any) {
		u.data.Value = nil
	})
}

func (u *ChatHistory) Len() int {
	// u.locker.RLock()
	// defer u.locker.RUnlock()
	return u.data.Len()
}

func (u *ChatHistory) Slice() []*llms.Message {
	// u.locker.RLock()
	// defer u.locker.RUnlock()
	x := make([]*llms.Message, 0, u.data.Len())
	u.data.Do(func(a any) {
		if a == nil {
			return
		}
		x = append(x, a.(*llms.Message))
	})
	return x
}

func (u *ChatHistory) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Slice())
}

func (u *ChatHistory) ToJSON() string {
	b, err := json.Marshal(u.Slice())
	if err != nil {
		return ""
	}
	return json.String(b)
}

func (u *ChatHistory) FromJSON(s string) error {
	a := make([]*llms.Message, 0)
	err := json.Unmarshal(json.Bytes(s), &a)
	if err != nil {
		return err
	}
	u.StoreMany(a...)
	return nil
}
