package crypto

import (
	"bytes"
	"testing"
)

func TestNewCompressor_GZip(t *testing.T) {
	c := NewCompressor(CompressGZip)
	if c == nil {
		t.Fatal("NewCompressor(CompressGZip) returned nil")
	}
	data := []byte("hello gzip compressor")
	enc, err := c.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if bytes.Equal(enc, data) {
		t.Error("Encoded data should not be equal to original")
	}
	dec, err := c.Decode(enc)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(dec, data) {
		t.Errorf("Decoded data mismatch: got %q, want %q", dec, data)
	}
}

func TestNewCompressor_Zlib(t *testing.T) {
	c := NewCompressor(CompressZlib)
	if c == nil {
		t.Fatal("NewCompressor(CompressZlib) returned nil")
	}
	data := []byte("hello zlib compressor")
	enc, err := c.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if bytes.Equal(enc, data) {
		t.Error("Encoded data should not be equal to original")
	}
	dec, err := c.Decode(enc)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(dec, data) {
		t.Errorf("Decoded data mismatch: got %q, want %q", dec, data)
	}
}

func TestNewCompressor_Zstd(t *testing.T) {
	c := NewCompressor(CompressZstd)
	if c == nil {
		t.Fatal("NewCompressor(CompressZstd) returned nil")
	}
	data := []byte("hello zstd compressor")
	enc, err := c.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if bytes.Equal(enc, data) {
		t.Error("Encoded data should not be equal to original")
	}
	dec, err := c.Decode(enc)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(dec, data) {
		t.Errorf("Decoded data mismatch: got %q, want %q", dec, data)
	}
}

func TestNewCompressor_InvalidType(t *testing.T) {
	// Test with an invalid type (not defined in consts)
	c := NewCompressor(CompressType(99))
	if c == nil {
		t.Fatal("NewCompressor(invalid type) returned nil")
	}
	data := []byte("test invalid type")
	_, err := c.Encode(data)
	if err == nil {
		t.Error("Expected error for invalid compressor type, got nil")
	}
}
