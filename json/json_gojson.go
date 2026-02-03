//go:build !amd64

package json

import (
	gojson "encoding/json"

	json "github.com/goccy/go-json"
)

var (
	// Valid 验证
	Valid = json.Valid
	// MarshalIndent 带缩进的序列化
	MarshalIndent = json.MarshalIndent

	Compact = gojson.Compact
)

// Marshal json.MarshalWithOption
func Marshal(v any) ([]byte, error) {
	return json.MarshalWithOption(v, json.UnorderedMap(), json.DisableNormalizeUTF8())
}

// MarshalToString json.MarshalWithOption and return string
func MarshalToString(v any) (string, error) {
	b, err := Marshal(v)
	if err == nil {
		return String(b), nil
	}
	return "", err
}
func Unmarshal(data []byte, v any) error {
	return json.UnmarshalWithOption(data, v, json.DecodeFieldPriorityFirstWin())
}

// UnmarshalFromString json.UnmarshalFromString
func UnmarshalFromString(data string, v any) error {
	return json.UnmarshalWithOption(Bytes(data), v, json.DecodeFieldPriorityFirstWin())
}

// import jsoniter "github.com/json-iterator/go"

// var (
// 	ijson = jsoniter.Config{
// 		OnlyTaggedField:               true,
// 		ObjectFieldMustBeSimpleString: true,
// 	}.Froze()

// 	Valid   = ijson.Valid
// 	Marshal = ijson.Marshal
// 	// Unmarshal is exported by gin/json package.
// 	Unmarshal = ijson.Unmarshal
// 	// MarshalIndent is exported by gin/json package.
// 	MarshalIndent = ijson.MarshalIndent
// 	// NewDecoder is exported by gin/json package.
// 	NewDecoder = ijson.NewDecoder
// 	// NewEncoder is exported by gin/json package.
// 	NewEncoder = ijson.NewEncoder
// )

// // MarshalToString json.MarshalWithOption and return string
// func MarshalToString(v any) (string, error) {
// 	b, err := Marshal(v)
// 	if err == nil {
// 		return String(b), nil
// 	}
// 	return "", err
// }

// // UnmarshalFromString json.UnmarshalFromString
// func UnmarshalFromString(data string, v any) error {
// 	return Unmarshal(Bytes(data), v)
// }
