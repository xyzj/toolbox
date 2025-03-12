package ollama

import (
	"bytes"
	"testing"
)

func TestOLL(t *testing.T) {
	c := NewChat(OptServerAddr("http://192.168.50.97:11434"), OptModel("gemma2"))
	buf := &bytes.Buffer{}
	err := c.Chat("golang有什么代码编写方法能在循环中优化内存性能", func(b []byte) error {
		buf.Write(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	println(buf.String())
}
