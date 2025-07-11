package config

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/json"
)

type dataType byte

const (
	tstr dataType = iota
	tint64
	tuint64
	tfloat64
	tbool
)

// EmptyValue an empty value
var EmptyValue = &Value{t: tstr}

// NewValue return a value
func NewValue(s string) *Value {
	return &Value{
		nstr: s,
	}
}

// NewInt64Value return a value
func NewInt64Value(n int64) *Value {
	return &Value{
		t:      tint64,
		nint64: n,
	}
}

// NewUint64Value return a value
func NewUint64Value(n uint64) *Value {
	return &Value{
		t:       tuint64,
		nuint64: n,
	}
}

// NewFloat64Value return a value
func NewFloat64Value(n float64) *Value {
	return &Value{
		t:        tfloat64,
		nfloat64: n,
	}
}

// NewBoolValue return a value
func NewBoolValue(n bool) *Value {
	return &Value{
		t:     tbool,
		nbool: n,
	}
}

// NewCodeValue return a value after code the data
func NewCodeValue(s string) *Value {
	return &Value{
		nstr: toolbox.CodeString(s),
	}
}

func NewObfuscatedValue(s string) *Value {
	return &Value{
		nstr: crypto.ObfuscationString(s),
	}
}

// Value config value
type Value struct {
	nstr     string
	nint64   int64
	nuint64  uint64
	nfloat64 float64
	nbool    bool
	t        dataType
}

func (v *Value) unmarshal() error {
	var err error
	v.nint64, err = strconv.ParseInt(v.nstr, 10, 64)
	if err == nil {
		v.t = tint64
		return nil
	}
	v.nuint64, err = strconv.ParseUint(v.nstr, 10, 64)
	if err == nil {
		v.t = tuint64
		return nil
	}
	v.nfloat64, err = strconv.ParseFloat(v.nstr, 64)
	if err == nil {
		v.t = tfloat64
		return nil
	}
	v.nbool, err = strconv.ParseBool(v.nstr)
	if err == nil {
		v.t = tbool
		return nil
	}
	return nil
}

func (v *Value) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&v.nstr); err != nil {
		return err
	}
	return v.unmarshal()
}

func (v *Value) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == 34 {
		data = data[1 : len(data)-1]
	}
	v.nstr = json.String(data)
	return v.unmarshal()
}

func (v *Value) MarshalYAML() (any, error) {
	switch v.t {
	case tint64:
		return v.nint64, nil
	case tuint64:
		return v.nuint64, nil
	case tfloat64:
		return v.nfloat64, nil
	case tbool:
		return v.nbool, nil
	default:
		return v.nstr, nil
	}
}

func (v *Value) MarshalJSON() ([]byte, error) {
	switch v.t {
	case tint64:
		return strconv.AppendInt([]byte{}, v.nint64, 10), nil // []byte(fmt.Sprintf("%d", v.nint64)), nil
	case tuint64:
		return strconv.AppendUint([]byte{}, v.nuint64, 10), nil // []byte(fmt.Sprintf("%d", v.nuint64)), nil
	case tfloat64:
		return strconv.AppendFloat([]byte{}, v.nfloat64, 'g', -1, 64), nil // []byte(fmt.Sprintf("%g", v.nfloat64)), nil
	case tbool:
		return strconv.AppendBool([]byte{}, v.nbool), nil // return []byte(fmt.Sprintf("%t", v.nbool)), nil
	default:
		return json.Bytes("\"" + strings.ReplaceAll(v.nstr, `"`, `\"`) + "\""), nil
	}
}

// String reutrn string
func (v *Value) String() string {
	switch v.t {
	case tstr:
		return v.nstr
	case tint64:
		v.nstr = strconv.FormatInt(v.nint64, 10)
	case tuint64:
		v.nstr = strconv.FormatUint(v.nuint64, 10)
	case tfloat64:
		v.nstr = strconv.FormatFloat(v.nfloat64, 'g', -1, 64)
	case tbool:
		v.nstr = strconv.FormatBool(v.nbool)
	}
	return v.nstr
}

// Bytes reutrn []byte
func (v *Value) Bytes() []byte {
	return json.Bytes(v.nstr)
}

// TryBool reutrn bool
func (v *Value) TryBool() bool {
	switch v.t {
	case tbool:
		return v.nbool
	case tint64:
		return v.nint64 != 0
	case tuint64:
		return v.nuint64 != 0
	case tfloat64:
		return v.nfloat64 != 0
	}
	var err error
	v.nbool, err = strconv.ParseBool(v.nstr)
	if err == nil {
		if v.t == tstr {
			v.t = tbool
		}
		return v.nbool
	}
	return false
}

// TryInt reutrn int
func (v *Value) TryInt() int {
	if v.t == tint64 {
		return int(v.nint64)
	}
	return int(v.TryInt64())
}

// TryInt32 reutrn int32
func (v *Value) TryInt32() int32 {
	if v.t == tint64 {
		return int32(v.nint64)
	}
	return int32(v.TryInt64())
}

// TryInt64 reutrn int64
func (v *Value) TryInt64() int64 {
	switch v.t {
	case tint64:
		return v.nint64
	case tuint64:
		return int64(v.nuint64)
	case tfloat64:
		return int64(v.nfloat64)
	case tbool:
		if v.nbool {
			return 1
		}
		return 0
	}
	var err error
	v.nint64, err = strconv.ParseInt(v.nstr, 10, 64)
	if err == nil {
		v.t = tint64
		return v.nint64
	}
	return 0
}

// TryUint64 reutrn uint64
func (v *Value) TryUint64() uint64 {
	switch v.t {
	case tuint64:
		return v.nuint64
	case tint64:
		return uint64(v.nint64)
	case tfloat64:
		return uint64(v.nfloat64)
	case tbool:
		if v.nbool {
			return 1
		}
		return 0
	}
	var err error
	v.nuint64, err = strconv.ParseUint(v.nstr, 10, 64)
	if err == nil {
		v.t = tuint64
		return v.nuint64
	}
	return 0
}

// TryFloat32 reutrn float32
// dec: 设置小数位数，仅第一个值有效
func (v *Value) TryFloat32(dec ...int) float32 {
	return float32(v.TryFloat64(dec...))
}

// TryFloat64 reutrn float64
// dec: 设置小数位数，仅第一个值有效
func (v *Value) TryFloat64(dec ...int) float64 {
	switch v.t {
	case tfloat64:
		if len(dec) == 0 {
			return v.nfloat64
		}
		x := math.Pow10(dec[0])
		return math.Trunc(v.nfloat64*x+0.5) / x
	case tuint64:
		return float64(v.nuint64)
	case tint64:
		return float64(v.nint64)
	case tbool:
		if v.nbool {
			return 1
		}
		return 0
	}
	var err error
	v.nfloat64, err = strconv.ParseFloat(v.nstr, 64)
	if err == nil {
		v.t = tfloat64
		return v.nfloat64
	}
	return 0
}

func (v *Value) TryDecode() string {
	if s := toolbox.DecodeString(v.nstr); s != "" {
		return s
	}
	return v.nstr
}

func (v *Value) TryDeobfuscation() string {
	if s := crypto.DeobfuscationString(v.nstr); s != "" {
		return s
	}
	return v.nstr
}

func (v *Value) TryTimestamp(f string) int64 {
	if f == "" {
		f = toolbox.DateTimeFormat
	}
	_, tz := time.Now().Zone()
	return toolbox.Time2Stampf(v.nstr, f, float32(tz)/3600)
}

// PwdString 序列化反序列化时可自动加密解密字符串，用于敏感字段
type PwdString string

func (p *PwdString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = PwdString(toolbox.DecodeString(s))
	return nil
}

func (p *PwdString) MarshalJSON() ([]byte, error) {
	if string(*p) == "" {
		return []byte("\"\""), nil
	}
	return []byte("\"" + toolbox.CodeString(string(*p)) + "\""), nil
}

func (p *PwdString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*p = PwdString(toolbox.DecodeString(s))
	return nil
}

func (p PwdString) MarshalYAML() (interface{}, error) {
	if string(p) == "" {
		return "", nil
	}
	return toolbox.CodeString(string(p)), nil
}

// VString value string, can parse to bool int64 float64
type VString string

// String reutrn string
func (rs VString) String() string {
	return string(rs)
}

// Bytes reutrn []byte
func (rs VString) Bytes() []byte {
	return []byte(rs)
}

// TryBool reutrn bool
func (rs VString) TryBool() bool {
	v, _ := strconv.ParseBool(string(rs))
	return v
}

// TryInt reutrn int
func (rs VString) TryInt() int {
	return toolbox.String2Int(string(rs), 10)
}

// TryInt32 reutrn int32
func (rs VString) TryInt32() int32 {
	return toolbox.String2Int32(string(rs), 10)
}

// TryInt64 reutrn int64
func (rs VString) TryInt64() int64 {
	return toolbox.String2Int64(string(rs), 10)
}

// TryUint64 reutrn uint64
func (rs VString) TryUint64() uint64 {
	return toolbox.String2UInt64(string(rs), 10)
}

// TryFloat32 reutrn float32
func (rs VString) TryFloat32() float32 {
	return toolbox.String2Float32(string(rs))
}

// TryFloat64 reutrn float64
func (rs VString) TryFloat64() float64 {
	return toolbox.String2Float64(string(rs))
}

// TryDecode try decode the value, if failed, return the origin
func (rs VString) TryDecode() string {
	if s := toolbox.DecodeString(string(rs)); s != "" {
		return s
	}
	return string(rs)
}

func (rs VString) TryDeobfuscation() string {
	if s := crypto.DeobfuscationString(string(rs)); s != "" {
		return s
	}
	return string(rs)
}

// TryTimestamp try turn time string to timestamp
//
// f: datetime format string，default is 2006-01-02 15:04:05, use timezone +8.0
func (rs VString) TryTimestamp(f string) int64 {
	if f == "" {
		f = toolbox.DateTimeFormat
	}
	return toolbox.Time2Stampf(string(rs), f, 8)
}
