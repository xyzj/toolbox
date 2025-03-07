package ollama

import (
	"testing"
	"time"
)

func TestOLL(t *testing.T) {
	println(idHash.Hash([]byte(time.Now().String())).HexString())
}
