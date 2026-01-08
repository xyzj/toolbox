// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build amd64

package json

import "github.com/bytedance/sonic"

var (
	json = sonic.Config{
		NoValidateJSONMarshaler: true,
		NoValidateJSONSkip:      true,
		NoEncoderNewline:        true,
		EncodeNullForInfOrNan:   true,
		NoNullSliceOrMap:        true,
	}.Froze()
	// Unmarshal is exported by gin/json package.
	Unmarshal = json.Unmarshal
	// MarshalIndent is exported by gin/json package.
	MarshalIndent = json.MarshalIndent
	// NewDecoder is exported by gin/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by gin/json package.
	NewEncoder = json.NewEncoder
	// Valid reports whether the provided byte slice is valid JSON.
	// It is an alias for json.Valid from the standard library.
	Valid = json.Valid
)

func Marshal(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err == nil {
		if len(b) <= 2 {
			return []byte{}, nil
		}
		return b, nil
	}
	return []byte{}, err
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
