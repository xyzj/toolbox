// Package crypto 加密解密
package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"io"
	mrand "math/rand"
	"os"
	"strings"

	"github.com/xyzj/deepcopy"
	"github.com/xyzj/toolbox/json"
)

var EmptyValue = CValue([]byte{})

// CValue 加密后的数据，可输出[]byte,hex string,base64string
type CValue []byte

// Len 加密结果长度
func (v CValue) Len() int {
	return len(v)
}

// Bytes 加密结果
func (v CValue) Bytes() []byte {
	return []byte(deepcopy.CopyAny[CValue](v))
}

// HexString 加密结果以hex字符串形式输出
func (v CValue) HexString() string {
	return hex.EncodeToString(v)
}

// Base64String 加密结果以标准base64字符串形式输出
func (v CValue) Base64String() string {
	return base64.StdEncoding.EncodeToString(v)
}

// Base64StringNoTail 加密结果以标准base64字符串形式输出，去除`=`
func (v CValue) Base64StringNoTail() string {
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(v), "=", "")
}

// URLBase64String 加密结果以URLbase64字符串形式输出
func (v CValue) URLBase64String() string {
	return base64.URLEncoding.EncodeToString(v)
}

var (
	md5hash    = NewHash(HashMD5)
	sha1hash   = NewHash(HashSHA1)
	sha256hash = NewHash(HashSHA256)
	sha512hash = NewHash(HashSHA512)
	sm3hash    = NewHash(HashSM3)
)

type Cryptor interface {
	GenerateKey(bits RSABits) (CValue, CValue, error)
	SetPublicKeyFromFile(keyPath string) error
	SetPublicKey(key string) error
	SetPrivateKeyFromFile(keyPath string) error
	SetPrivateKey(key string) error
	Encode(b []byte) (CValue, error)
	Decode(b []byte) (string, error)
	DecodeBase64(s string) (string, error)
	Decrypt(s string) string
	Encrypt(s string) string
	EncryptTo(s string) CValue
}
type CertOpt struct {
	// 证书包含的域名清单
	DNS []string `json:"dns"`
	// 证书包含的ip清单
	IP []string `json:"ip"`
	// 根证书私钥，未指定或载入错误时，会重新生成私钥和根证书
	RootKey string `json:"root-key"`
	// 根证书，当私钥配置错误时，该参数无效
	RootCa string `json:"root-ca"`
	// 输出目录
	OutPut string `json:"-"`
	// 组织
	Organization string `json:"organization"`
	// 国家
	Country string `json:"country"`
	// 省份
	Province string `json:"province"`
	// 城市
	Locality string `json:"locality"`
}

// GetMD5 生成md5字符串
func GetMD5(text string) string {
	return md5hash.Hash(json.Bytes(text)).HexString()
}

// GetSHA1 生成sha1字符串
func GetSHA1(text string) string {
	return sha1hash.Hash(json.Bytes(text)).HexString()
}

// GetSHA256 生成sha256字符串
func GetSHA256(text string) string {
	return sha256hash.Hash(json.Bytes(text)).HexString()
}

// GetSHA512 生成sha512字符串
func GetSHA512(text string) string {
	return sha512hash.Hash(json.Bytes(text)).HexString()
}

// GetSM3 生成sm3字符串
func GetSM3(text string) string {
	return sm3hash.Hash(json.Bytes(text)).HexString()
}

// GetRandom 获取随机数据
func GetRandom(l int) []byte {
	if l == 0 {
		return []byte{}
	}
	buf := make([]byte, l)
	io.ReadFull(rand.Reader, buf)
	return buf
}

// TLSConfigFromFile 从文件载入证书
func TLSConfigFromFile(certfile, keyfile, rootfile string) (*tls.Config, error) {
	bcert, err := os.ReadFile(certfile)
	if err != nil {
		return nil, err
	}
	bkey, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	broot, _ := os.ReadFile(rootfile)
	return TLSConfigFromPEM(bcert, bkey, broot)
}

// TLSConfigFromPEM 从pem载入证书
func TLSConfigFromPEM(certpem, keypem, rootpem []byte) (*tls.Config, error) {
	cliCrt, err := tls.X509KeyPair(certpem, keypem)
	if err != nil {
		return nil, err
	}
	tc := &tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		Certificates: []tls.Certificate{cliCrt},
	}
	if len(rootpem) == 0 {
		return tc, nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	tc.ClientCAs = pool
	if pool.AppendCertsFromPEM(rootpem) {
		tc.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tc, nil
}

func ObfuscationString(s string) string {
	x := byte(mrand.Int31n(126) + 1)
	ll := len(s)
	l := mrand.Intn(7) + 2
	if ll <= l {
		l = 2
	}
	salt := GetRandom(ll)
	y := bytes.Buffer{}
	y.Grow(ll * 2)
	y.WriteByte(x)
	y.WriteByte(byte(l))
	y.WriteByte(salt[0])
	c1 := 1
	k := 0
	ss := json.Bytes(s)
	for _, v := range ss {
		if k == l {
			y.WriteByte(salt[c1])
			c1++
			k = 1
		} else {
			k++
		}
		y.WriteByte(v + x)
	}
	y.Write(GetRandom(3))
	zz := y.Bytes()
	zz = json.ReverseBytes(zz)
	a := base64.URLEncoding.EncodeToString(zz)
	a = json.ReverseString(a)
	a = json.SwapCase(a)
	a = strings.Replace(a, "=", "", -1)
	return a
}

func DeobfuscationString(s string) string {
	s = FillBase64(json.ReverseString(json.SwapCase(s)))
	a, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	a = json.ReverseBytes(a)
	x := a[0]
	l := int(a[1])
	a = a[3 : len(a)-3]
	// buf := bytes.Buffer{}
	// for _, v := range a {
	// 	buf.WriteByte(v)
	// }
	y := bytes.Buffer{}
	y.Grow(len(a))
	k := 0
	for _, v := range a {
		if k == l {
			k = 0
			continue
		}
		k++
		y.WriteByte(v - x)
	}
	return y.String()
}
