package json

import (
	"bytes"
	"strings"
	"unicode"
	"unsafe"
)

// ASCII 过滤非ASCII字符
func ASCII(b []byte) string {
	result := make([]byte, 0, len(b))
	for _, b1 := range b {
		// ASCII 标准范围是 0-127
		// 如果要排除 \x00 (NUL) 等控制字符，可以判断 b >= 32 && b <= 126
		if b1 >= 32 && b1 <= 126 {
			result = append(result, b1)
		}
	}
	return String(result)
}

// Bytes 内存地址转换string
func Bytes(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// String 内存地址转换[]byte
func String(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// PB2Json pb2格式转换为json []byte格式
func PB2Json(pb any) []byte {
	jsonBytes, err := Marshal(pb)
	if err != nil {
		return nil
	}
	return jsonBytes
}

// PB2String pb2格式转换为json 字符串格式
func PB2String(pb any) string {
	b, err := MarshalToString(pb)
	if err != nil {
		return ""
	}
	return b
}

// JSON2PB json字符串转pb2格式
func JSON2PB(js string, pb any) error {
	err := Unmarshal(Bytes(js), &pb)
	return err
}

// SwapCase swap char case
func SwapCase(s string) string {
	var ns bytes.Buffer
	for _, v := range s {
		if v >= 65 && v <= 90 {
			ns.WriteString(string(v + 32))
		} else if v >= 97 && v <= 122 {
			ns.WriteString(string(v - 32))
		} else {
			ns.WriteString(string(v))
		}
	}
	return ns.String()
}

// RemoveUnvisiable removes non-printable and symbol characters from a string.
func RemoveUnvisiable(s string) string {
	buf := strings.Builder{}
	buf.Grow(len(s))
	for _, r := range s {
		if unicode.IsPrint(r) && !unicode.Is(unicode.So, r) {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// ReverseString reverse string in place
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// ReverseBytes reverse byte slice in place
func ReverseBytes(b []byte) []byte {
	for from, to := 0, len(b)-1; from < to; from, to = from+1, to-1 {
		b[from], b[to] = b[to], b[from]
	}
	return b
}
