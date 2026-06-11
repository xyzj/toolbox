package toolbox

import (
	"testing"
)

func TestCode(t *testing.T) {
	// b := []byte{0x42, 0x48, 0, 0}
	// b := []byte{0, 0, 0x42, 0x48}
	// b := []byte{0x41, 0x13, 0x33, 0x33}
	// b := []byte{0, 0, 0x48, 0x14}
	// b := []byte{0x42, 0xb8, 0x73, 0x33}
	b := Float32ToByteBig(50.0)
	println(Bytes2String(b, "-"))
}
