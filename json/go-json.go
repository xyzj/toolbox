//go:build !sonic_json

package json

import (
	json "github.com/goccy/go-json"
)

var (
	// Valid 验证
	Valid = json.Valid
	// Unmarshal 反序列化
	Unmarshal = json.Unmarshal
	// MarshalIndent 带缩进的序列化
	MarshalIndent = json.MarshalIndent
)

// Marshal json.MarshalWithOption
func Marshal(v interface{}) ([]byte, error) {
	return json.MarshalWithOption(v, json.UnorderedMap())
}

// MarshalToString json.MarshalWithOption and return string
func MarshalToString(v interface{}) (string, error) {
	b, err := Marshal(v)
	if err == nil {
		return String(b), nil
	}
	return "", err
}

// UnmarshalFromString json.UnmarshalFromString
func UnmarshalFromString(data string, v interface{}) error {
	return Unmarshal(Bytes(data), v)
}
