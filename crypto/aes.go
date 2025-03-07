package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/xyzj/toolbox/json"
)

type AESType byte

const (
	// AES128CBC aes128cbc算法
	AES128CBC AESType = iota
	// AES192CBC aes192cbc算法
	AES192CBC
	// AES256CBC aes256cbc算法
	AES256CBC
	// AES128CFB aes128cfb算法
	AES128CFB
	// AES192CFB aes192cfb算法
	AES192CFB
	// AES256CFB aes256cfb算法
	AES256CFB
	// AES128ECB aes128ecb算法
	AES128ECB
	// AES192ECB aes192ecb算法
	AES192ECB
	// AES256ECB aes256ecb算法
	AES256ECB
)

// AES aes算法
type AES struct {
	padding   func(ciphertext []byte, blockSize int) []byte
	unpadding func(encrypt []byte) []byte
	block     cipher.Block
	iv        []byte
	blockSize int
	workType  AESType
	appendiv  bool
}

// SetPadding 设置填充模式
func (w *AES) SetPadding(p Padding) {
	switch p {
	case Pkcs5Padding, Pkcs7Padding:
		w.padding = pkcs7Padding
		w.unpadding = pkcs7Unpadding
	case ZeroPadding:
		w.padding = zeroPadding
		w.unpadding = zeroUnPadding
	default:
		w.padding = noPadding
		w.unpadding = noUnpadding
	}
}

// SetKeyIV 设置iv和key
// 如果不设置iv，会生成随机iv并追加在加密结果的头部
func (w *AES) SetKeyIV(key, iv string) error {
	l := 16
	switch w.workType {
	case AES192CBC, AES192CFB, AES192ECB:
		l = 24
	case AES256CBC, AES256CFB, AES256ECB:
		l = 32
	}
	if len(key) < l {
		return fmt.Errorf("key length must be longer than %d", l)
	}
	biv := json.Bytes(iv)
	switch w.workType {
	case AES128ECB, AES192ECB, AES256ECB:
	default:
		if len(biv) == 0 {
			w.appendiv = true
			biv = GetRandom(aes.BlockSize)
		}
		if len(iv) < aes.BlockSize {
			return fmt.Errorf("the length of iv must be longer than %d", aes.BlockSize)
		}
		w.iv = biv[:aes.BlockSize]
	}
	w.block, _ = aes.NewCipher(json.Bytes(key[:l]))
	w.blockSize = w.block.BlockSize()
	return nil
}

// Encode aes加密
func (w *AES) Encode(b []byte) (CValue, error) {
	if w.block == nil {
		return EmptyValue, fmt.Errorf("key or iv are not set")
	}
	content := w.padding(b, aes.BlockSize)
	var crypted []byte
	idx := 0
	if w.appendiv {
		crypted = make([]byte, aes.BlockSize+len(content))
		copy(crypted, w.iv)
		idx = aes.BlockSize
	} else {
		crypted = make([]byte, len(content))
	}
	// 不能预初始化，否则二次编码过程会出错
	switch w.workType {
	case AES128CBC, AES192CBC, AES256CBC:
		cipher.NewCBCEncrypter(w.block, w.iv).CryptBlocks(crypted[idx:], content)
	case AES128CFB, AES192CFB, AES256CFB:
		cipher.NewCFBEncrypter(w.block, w.iv).XORKeyStream(crypted[idx:], content)
	case AES128ECB, AES192ECB, AES256ECB:
		for bs, be := 0, w.blockSize; bs < len(content); bs, be = bs+w.blockSize, be+w.blockSize {
			w.block.Encrypt(crypted[bs:be], content[bs:be])
		}
	}
	return CValue(crypted), nil
}

// Decode aes解密
func (w *AES) Decode(b []byte) (string, error) {
	if w.block == nil {
		return "", fmt.Errorf("key or iv are not set")
	}
	if w.appendiv {
		w.iv = b[:aes.BlockSize]
		b = b[aes.BlockSize:]
	}
	switch w.workType {
	case AES128CBC, AES192CBC, AES256CBC:
		decrypted := make([]byte, len(b))
		cipher.NewCBCDecrypter(w.block, w.iv).CryptBlocks(decrypted, b)
		return json.String(pkcs7Unpadding(decrypted)), nil
	case AES128CFB, AES192CFB, AES256CFB:
		cipher.NewCFBDecrypter(w.block, w.iv).XORKeyStream(b, b)
		return json.String(w.unpadding(b)), nil
	case AES128ECB, AES192ECB, AES256ECB:
		decrypted := make([]byte, len(b))
		for bs, be := 0, w.blockSize; bs < len(b); bs, be = bs+w.blockSize, be+w.blockSize {
			w.block.Decrypt(decrypted[bs:be], b[bs:be])
		}
		return json.String(w.unpadding(decrypted)), nil
	}
	return "", fmt.Errorf("unsupport cipher type")
}

// DecodeBase64 aes解密base64编码的字符串
func (w *AES) DecodeBase64(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(FillBase64(s))
	if err != nil {
		return "", err
	}
	return w.Decode(b)
}

// Decrypt 兼容旧方法，直接解析base64字符串
func (w *AES) Decrypt(s string) string {
	x, _ := w.DecodeBase64(s)
	return x
}

// Encrypt 兼容旧方法，直接返回base64字符串
func (w *AES) Encrypt(s string) string {
	x, err := w.Encode(json.Bytes(s))
	if err != nil {
		return ""
	}
	return x.Base64String()
}

// EncryptTo 兼容旧方法，直接返回base64字符串
func (w *AES) EncryptTo(s string) CValue {
	x, err := w.Encode(json.Bytes(s))
	if err != nil {
		return EmptyValue
	}
	return x
}

// NewAES 创建一个新的aes加密解密器
func NewAES(t AESType) *AES {
	w := &AES{
		workType: t,
	}
	switch t {
	case AES128CBC, AES192CBC, AES256CBC, AES128ECB, AES192ECB, AES256ECB:
		w.SetPadding(Pkcs7Padding)
	default:
		w.SetPadding(NoPadding)
	}
	return w
}
