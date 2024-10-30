package coord

import "testing"

func TestBaidu(t *testing.T) {
	a := &Point{
		Lng: 116.677569,
		Lat: 23.36152191,
	}
	b := WGS84toBD09(a)
	println(b.String())
}

func TestGao(t *testing.T) {
	a := &Point{
		Lat: 24.774268771530696,
		Lng: 115.0237448410683,
	}
	println(WGS84toGCJ02(a).String())
}
