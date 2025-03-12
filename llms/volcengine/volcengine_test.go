package volcengine

import (
	"bytes"
	"testing"
)

func TestChatVol(t *testing.T) {
	c := NewChat(OptAPIKey("64095d54-432d-493c-be3b-8b1c06e8b9ee"), OptModel("doubao-1-5-lite-32k-250115"))
	buf := &bytes.Buffer{}
	err := c.Chat("golang有什么代码编写方法能在循环中优化内存性能", func(b []byte) error {
		buf.Write(b)
		return nil
	})
	if err != nil {
		println("Error")
		t.Fatal(err)
	}
	println(buf.String())
}
