package mapfx

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/xyzj/toolbox/coord"
)

func GetRandomString(l int64, letteronly ...bool) string {
	str := "!#%&()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}"
	if len(letteronly) > 0 && letteronly[0] {
		str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	bb := []byte(str)
	var rs strings.Builder
	for i := int64(0); i < l; i++ {
		rs.WriteByte(bb[rand.Intn(len(bb))])
	}
	return rs.String()
}

func TestSlice(t *testing.T) {
	ts := NewSliceMap[string]()
	for i := 0; i < 10; i++ {
		ts.StoreItem("未知区域", GetRandomString(10, true))
	}
	b := ts.Clone()
	println(fmt.Sprintf("%+v", b))
}

func TestEqual(t *testing.T) {
	a := &coord.Point{
		Lng: 123.456,
		Lat: 456.789,
	}
	b := &coord.Point{
		Lng: 123.456,
		Lat: 456.789,
	}
	println(a, b)
	println(a == b)
	println(reflect.DeepEqual(a, b))
}
