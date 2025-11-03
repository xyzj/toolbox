package cache

import (
	"crypto/rand"
	"fmt"
)

// 常用字符集常量
const (
	CharsetDigits  = "0123456789"
	CharsetLower   = "abcdefghijklmnopqrstuvwxyz"
	CharsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetSpecial = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

// RandomStringFromCharset 使用指定的字符集生成指定长度的随机字符串。
// 使用 crypto/rand 生成随机数并通过拒绝采样（rejection sampling）避免偏差。
// 如果 charset 为空或 length < 0 则返回错误。
func RandomStringFromCharset(charset string, length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("invalid length: %d", length)
	}
	if length == 0 {
		return "", nil
	}
	if charset == "" {
		return "", fmt.Errorf("empty charset")
	}

	n := len(charset)
	if n == 0 {
		return "", fmt.Errorf("empty charset")
	}

	// result buffer
	out := make([]byte, length)

	// max acceptable random byte to avoid modulo bias:
	// accept byte values in [0, maxAccept], where maxAccept+1 is a multiple of n
	maxAccept := 255 - (256 % n)

	// temp buffer read size (read in chunks)
	buf := make([]byte, 64)

	i := 0
	for i < length {
		_, err := rand.Read(buf)
		if err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) > maxAccept {
				continue
			}
			out[i] = charset[int(b)%n]
			i++
			if i >= length {
				break
			}
		}
	}

	return string(out), nil
}

// RandomStringFromSets 将多个字符集合并后生成随机字符串（方便传入多个常量）。
// 例如： RandomStringFromSets([]string{CharsetDigits, CharsetLower}, 16)
func RandomStringFromSets(sets []string, length int) (string, error) {
	if len(sets) == 0 {
		return "", fmt.Errorf("no charset provided")
	}
	// 合并字符集（简单拼接，不去重）
	totalLen := 0
	for _, s := range sets {
		totalLen += len(s)
	}
	if totalLen == 0 {
		return "", fmt.Errorf("combined charset empty")
	}
	merged := make([]byte, 0, totalLen)
	for _, s := range sets {
		merged = append(merged, s...)
	}
	return RandomStringFromCharset(string(merged), length)
}
