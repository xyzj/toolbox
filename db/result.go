package db

import (
	"encoding/hex"
	"hash/crc32"
	"strconv"
	"sync"
	"time"

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

const (
	defaultDateTimeFormat = "2006-01-02 15:04:05"
)

var EmptyValue = Value{}

type Value struct {
	val any
}

// TryTime returns a formatted time string based on the underlying value and format.
func (v *Value) TryTime(f string) string {
	if v == nil {
		return ""
	}
	if f == "" {
		f = defaultDateTimeFormat
	}
	switch b := v.val.(type) {
	case time.Time:
		return b.Format(f)
	case int64:
		return time.Unix(b, 0).Format(f)
	case uint64:
		return time.Unix(int64(b), 0).Format(f)
	case float64:
		return time.Unix(int64(b), 0).Format(f)
	case float32:
		return time.Unix(int64(b), 0).Format(f)
	case []uint8:
		t, err := time.Parse(f, json.String(b))
		if err != nil {
			return ""
		}
		return t.Format(f)
	default:
		return ""
	}
}

// TryTimestamp returns a unix timestamp derived from the underlying value.
func (v *Value) TryTimestamp(f string) int64 {
	if v == nil {
		return 0
	}
	switch b := v.val.(type) {
	case time.Time:
		return b.Unix()
	case int64:
		return b
	case uint64:
		return int64(b)
	case float64:
		return int64(b)
	case float32:
		return int64(b)
	case []uint8:
		if f == "" {
			f = defaultDateTimeFormat
		}
		t, err := time.Parse(f, json.String(b))
		if err != nil {
			return 0
		}
		return t.Unix()
	default:
		return 0
	}
}

// String returns the string representation of the underlying value.
func (v *Value) String() string {
	if v == nil {
		return ""
	}
	switch b := v.val.(type) {
	case []uint8:
		return json.String(b)
	case float32:
		return strconv.FormatFloat(float64(b), 'f', -1, 32)
	case int64:
		return strconv.FormatInt(b, 10)
	case uint64:
		return strconv.FormatUint(b, 10)
	case float64:
		return strconv.FormatFloat(b, 'f', -1, 64)
	case bool:
		if b {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// TryInt64 converts the underlying value to int64 if possible.
func (v *Value) TryInt64() int64 {
	if v == nil {
		return 0
	}
	switch b := v.val.(type) {
	case int64:
		return b
	case uint64:
		return int64(b)
	case float64:
		return int64(b)
	case float32:
		return int64(b)
	case []uint8:
		i, _ := strconv.ParseInt(json.String(b), 10, 64)
		return i
	case bool:
		if b {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// TryFloat64 converts the underlying value to float64 with optional precision.
func (v *Value) TryFloat64(dec ...int) float64 {
	if v == nil {
		return 0
	}
	xdec := 2
	if len(dec) > 0 && dec[0] > 0 {
		xdec = dec[0]
	}
	var val float64
	switch b := v.val.(type) {
	case float64:
		val = b
	case float32:
		val = float64(b)
	case []uint8:
		val, _ = strconv.ParseFloat(json.String(b), 64)
	case int64:
		val = float64(b)
	case uint64:
		val = float64(b)
	case bool:
		if b {
			val = 1
		}
	default:
	}
	s := strconv.FormatFloat(val, 'f', xdec, 64)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// TryBool converts the underlying value to bool if possible.
func (v *Value) TryBool() bool {
	if v == nil {
		return false
	}
	switch b := v.val.(type) {
	case bool:
		return b
	case int64:
		return b != 0
	case uint64:
		return b != 0
	case float64:
		return b != 0
	case float32:
		return b != 0
	case []uint8:
		s := json.String(b)
		return s == "true" || s == "1"
	default:
		return false
	}
}

// TryUint64 converts the underlying value to uint64 if possible.
func (v *Value) TryUint64() uint64 {
	if v == nil {
		return 0
	}
	switch b := v.val.(type) {
	case uint64:
		return b
	case int64:
		return uint64(b)
	case float64:
		return uint64(b)
	case float32:
		return uint64(b)
	case []uint8:
		i, _ := strconv.ParseUint(json.String(b), 10, 64)
		return i
	case bool:
		if b {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// TryInt32 converts the underlying value to int32.
func (v *Value) TryInt32() int32 {
	return int32(v.TryInt64())
}

// TryInt converts the underlying value to int.
func (v *Value) TryInt() int {
	return int(v.TryInt64())
}

// TryFloat32 converts the underlying value to float32 with optional precision.
func (v *Value) TryFloat32(dec ...int) float32 {
	return float32(v.TryFloat64(dec...))
}

// QueryDataChan chan方式返回首页数据
type QueryDataChan struct {
	Locker *sync.WaitGroup
	Data   *QueryData
	Total  *int
	Err    error
}

func newDataRow(cols int) QueryDataRow {
	return QueryDataRow{
		Cells:  make([]string, cols),
		VCells: make([]Value, cols),
	}
}

// QueryDataRow 数据行
type QueryDataRow struct {
	// Deprecated: will removed in a future version, use VCells
	Cells  []string `json:"cells,omitempty"`
	VCells []Value  `json:"vcells,omitempty"`
}

func (d *QueryDataRow) JSON() string {
	s, _ := json.MarshalToString(d)
	return s
}

// QueryData 数据集
type QueryData struct {
	Rows     []QueryDataRow `json:"rows,omitempty"`
	Columns  []string       `json:"columns,omitempty"`
	CacheTag string         `json:"cache_tag,omitempty"`
	Total    int            `json:"total,omitempty"`
}

func (d *QueryData) JSON() (string, error) {
	return json.MarshalToString(d)
	// return s
}

func makeCacheTag(cachehead string) string {
	v := crc32.ChecksumIEEE(json.Bytes(time.Now().String()))
	var sumBuf [4]byte
	sumBuf[0] = byte(v >> 24)
	sumBuf[1] = byte(v >> 16)
	sumBuf[2] = byte(v >> 8)
	sumBuf[3] = byte(v)
	return cachehead + hex.EncodeToString(sumBuf[:])
}
