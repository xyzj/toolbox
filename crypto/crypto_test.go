package crypto

import (
	"testing"

	"github.com/xyzj/toolbox/json"
)

func TestHash(t *testing.T) {
	v := "kjhfksdfh2983u92fsdkfhakjdhf92837@#$^&*()"
	t.Run("hash md5", func(t *testing.T) {
		println(GetMD5(v))
	})
	t.Run("hash sha256", func(t *testing.T) {
		println(GetSHA256(v))
	})
	t.Run("hash sha512", func(t *testing.T) {
		println(GetSHA512(v))
	})
	t.Run("hash sha1", func(t *testing.T) {
		c := NewHash(HashSHA1, nil)
		println(c.Hash(json.Bytes(v)).HexString())
	})
}
