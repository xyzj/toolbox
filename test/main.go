package main

import (
	"path/filepath"
	"unsafe"
)

func toRawBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return []byte(s)
}
func toRawString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return string(b)
}
func toReflectBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
func toReflectString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&b))
}
func toBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
func toString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
func toK8sBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
func toK8sString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func main() {
	println(filepath.FromSlash(`some/path/to/file//`))
}
