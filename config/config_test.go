package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	a  = `{"dsde":de}}"""""""""""""""{34"4f.di34fs"4""""""""""""""""""3""""""""""""""""""\f3/daz!@#$"%^+_)(*&)}`
	bb = false
	c  = uint64(86233243234)
	d  = int64(1723512648)
	e  = float64(15563312.2556218585124577)
)

func TestTime(t *testing.T) {
	a, b := time.Now().Zone()
	println(a, b, float32(b)/60/60)
}

type aaa struct {
	A32 float32
	A64 float64
}

var (
	z  = float64(1235.2154678513156)
	aa *aaa
)

func BenchmarkSconv(b *testing.B) {
	a, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", z), 64)
	aa = &aaa{A32: float32(a), A64: a}
}

func BenchmarkStrQuote(b *testing.B) {
	a := math.Trunc(z*math.Pow10(5)+0.5) / math.Pow10(5)
	aa = &aaa{
		A64: a,
		A32: float32(a),
	}
}

func BenchmarkStrRep(b *testing.B) {
	strings.ReplaceAll(a, `"`, `\"`)
}

func BenchmarkConvFmt(b *testing.B) {
	fmt.Sprintf("%d", c)
	fmt.Sprintf("%d", d)
	fmt.Sprintf("%t", bb)
	fmt.Sprintf("%g", e)
}

func BenchmarkConvStr(b *testing.B) {
	strconv.AppendBool([]byte{}, bb)
	strconv.AppendFloat([]byte{}, e, 'g', -1, 64)
	strconv.AppendInt([]byte{}, d, 10)
	strconv.AppendUint([]byte{}, c, 10)
}

func TestConv(t *testing.T) {
	a, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", z), 64)
	aa = &aaa{A32: float32(a), A64: a}
	println(fmt.Sprintf("%+v", aa))
	a = math.Trunc(z*math.Pow10(0)+0.5) / math.Pow10(0)
	aa = &aaa{
		A64: a,
		A32: float32(a),
	}
	println(fmt.Sprintf("%+v", aa))
}

func TestConf(t *testing.T) {
	a := NewConfig("a.conf")
	// a.FromFile("")
	a.PutItem(&Item{Key: "key1", Value: NewValue("1231dsdf"), Comment: ""})
	a.PutItem(&Item{Key: "key2", Value: NewValue("adf4s"), Comment: ""})
	a.PutItem(&Item{Key: "key3", Value: NewValue("dfs3fg"), Comment: ""})
	a.PutItem(&Item{Key: "key4", Value: NewValue("vv323f"), Comment: ""})
	b := a.GetItem("key1")
	println(b.String())
	a.Save()
}
