package crypto

import (
	"strconv"
	"strings"
	"testing"
)

var prikey = "32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0"

// SplitStringWithLen 按制定长度分割字符串
//
//	s-原始字符串
//	l-切割长度
func SplitStringWithLen(s string, l int) []string {
	rs := []rune(s)
	ss := make([]string, 0)
	xs := ""
	for k, v := range rs {
		xs = xs + string(v)
		if (k+1)%l == 0 {
			ss = append(ss, xs)
			xs = ""
		}
	}
	if len(xs) > 0 {
		ss = append(ss, xs)
	}
	return ss
}

func String2Bytes(data string, sep string) []byte {
	var z []byte
	a := strings.Split(data, sep)
	z = make([]byte, len(a))
	for k, v := range a {
		b, _ := strconv.ParseUint(v, 16, 8)
		z[k] = byte(b)
	}
	return z
}

func TestSma2(t *testing.T) {
	// ss := SplitStringWithLen(prikey, 2)
	// xss := strings.Join(ss, "-")
	// bb := String2Bytes(xss, "-")
	// xb := base64.StdEncoding.EncodeToString(bb)
	sss := "1267312shfskdfadfaf"
	c := NewSM2()
	c.GenerateKey()
	c.ToFile("sm2pub.pem", "sm2pri.pem")
	err := c.SetPublicKeyFromFile("sm2pub.pem")
	if err != nil {
		t.Fatal(err)
		return
	}
	err = c.SetPrivateKeyFromFile("sm2pri.pem")
	if err != nil {
		t.Fatal(err)
		return
	}
	// c.ToFile("sm2pub.pem", "sm2pri.pem")
	// c.SetPrivateKey(xb)
	// c.SetPublicKey("BPWIR6XwO8j41zr153bbudpdwLQlIPrcE/PhiKeSx/ENrMJP4lYB7cGTEDBv0ZNdGvYdZrvN2VXBNTzkK91ASVQ=")
	s, err := c.Encode([]byte(sss))
	if err != nil {
		t.Fatal(err)
		return
	}
	println("hex", s.HexString())
	println("b64", s.Base64String())
	// b, _ := base64.StdEncoding.DecodeString("BG9efyZL93BZOxsHtNoaMACui4ZyMXm9Cjsq0xiQ4gFahCveqCE2fLQyboHZr/v6cy4i30nwNhh0UdgcUfe7xal4v53EnrAxHuRF/WjnWNborpAuMfl8FTfBPPsbPxdBJE3A4Jciztpmctvj4URDgw09yfo=")
	// s1, err := c.Decode(b)
	s1, err := c.Decode(s.Bytes())
	if err != nil {
		t.Fatal(err)
		return
	}
	if s1 != sss {
		t.Fail()
	}
	println(s1)
}

func TestSM4(t *testing.T) {
	sss := "1267312shfskdfadfaf"
	key, iv := []byte("nNZT3xhtcKyykgBtsn7OAx0cymmNEPqE"), []byte("4qzB9DK6eFuSOMfB")
	c := NewSM4(SM4CBC)
	c.SetKeyIV(key, iv)
	s, err := c.Encode([]byte(sss))
	if err != nil {
		t.Fail()
		return
	}
	s1, err := c.Decode(s.Bytes())
	if err != nil {
		t.Fail()
		return
	}
	if s1 != sss {
		t.Fail()
	}
}

func BenchmarkSM2(b *testing.B) {
	c := NewSM2()
	c.GenerateKey()
	// c.SetPublicKeyFromFile("publicKey.pem")
	// c.SetPrivateKeyFromFile("privateKey.pem")
	sss := `{"deviceCode":"2","kIndexes":[{"loopIndex":"K1","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K2","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K3","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]},{"loopIndex":"K4","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K5","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K6","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K7","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K8","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]}]}]}` // toolbox.GetRandomString(30002, true) // "1267312shfskdfadfaf"
	bb := []byte(sss)
	var err error
	var x CValue
	// var xs string
	x, _ = c.Encode(bb)
	bbb := x.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = c.Decode(bbb)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}
