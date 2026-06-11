package toolbox

import "testing"

func TestCheckIP(t *testing.T) {
	// s := "feee::38:fe9b:dc6b:fa25:b590"
	s := "[240e:688:200:b35:d9e5:13af:d30a:420d]"
	x, err := IPv6ToInt32Segments(s)
	if err != nil {
		t.Fatal(err)
	}
	println(Bytes2String(x, "-"))
	y, err := Int32SegmentsToIPv6(x)
	if err != nil {
		t.Fatal(err)
	}
	println(y)
}
