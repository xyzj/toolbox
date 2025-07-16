package json

import jsoniter "github.com/json-iterator/go"

// var (
// 	// Valid 验证
// 	Valid = json.Valid
// 	// Unmarshal 反序列化
// 	Unmarshal = json.Unmarshal
// 	// MarshalIndent 带缩进的序列化
// 	MarshalIndent = json.MarshalIndent
// )

// // Marshal json.MarshalWithOption
// func Marshal(v interface{}) ([]byte, error) {
// 	return json.MarshalWithOption(v, json.UnorderedMap())
// }

// // MarshalToString json.MarshalWithOption and return string
// func MarshalToString(v interface{}) (string, error) {
// 	b, err := Marshal(v)
// 	if err == nil {
// 		return String(b), nil
// 	}
// 	return "", err
// }

// // UnmarshalFromString json.UnmarshalFromString
// func UnmarshalFromString(data string, v interface{}) error {
// 	return Unmarshal(Bytes(data), v)
// }

var (
	ijson = jsoniter.ConfigFastest

	Valid = ijson.Valid
	// Unmarshal is exported by gin/json package.
	Unmarshal = ijson.Unmarshal
	// MarshalIndent is exported by gin/json package.
	MarshalIndent = ijson.MarshalIndent
	// NewDecoder is exported by gin/json package.
	NewDecoder = ijson.NewDecoder
	// NewEncoder is exported by gin/json package.
	NewEncoder = ijson.NewEncoder
)

// Marshal json.MarshalWithOption
func Marshal(v any) ([]byte, error) {
	return ijson.Marshal(v)
}

// MarshalToString json.MarshalWithOption and return string
func MarshalToString(v any) (string, error) {
	b, err := Marshal(v)
	if err == nil {
		return String(b), nil
	}
	return "", err
}

// UnmarshalFromString json.UnmarshalFromString
func UnmarshalFromString(data string, v any) error {
	return Unmarshal(Bytes(data), v)
}
