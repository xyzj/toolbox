package crypto

import (
	"testing"
)

var s = "toolbox.GetRandomString(4096, true)"

func TestAES(t *testing.T) {
	ae := NewAES(AES128CBC)
	ae.SetKeyIV(string(GetRandom(32)), string(GetRandom(16)))
	enc, err := ae.Encode([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("enc: %s", enc)
	dec, err := ae.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("dec: %s", dec)
	if string(dec) != s {
		t.Fatalf("aes decode error")
	}
}
