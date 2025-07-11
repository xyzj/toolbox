package toolbox

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GbkToUtf8 gbk编码转utf8
func GbkToUtf8(s []byte) ([]byte, error) {
	if utf8.Valid(s) {
		return s, nil
	}
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return s, e
	}
	return d, nil
}

// Utf8ToGbk utf8编码转gbk
func Utf8ToGbk(s []byte) ([]byte, error) {
	// if !isUtf8(s) {
	// 	return s, nil
	// }
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return s, e
	}
	return d, nil
}

// Float32ToByte 32位浮点转bytes
func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

// Float64ToByte 64位浮点转bytes
func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}

// Float32ToByte 32位浮点转bytes
func Float32ToByteBig(float float32) []byte {
	bits := math.Float32bits(float)

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, bits)

	return bytes
}

// Float64ToByte 64位浮点转bytes
func Float64ToByteBig(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)

	return bytes
}

// FormatFloat64 格式化浮点精度，f-浮点数，p-小数位数
func FormatFloat64(f float64, p int) float64 {
	// println(fmt.Sprintf("%.10f", f))
	x := math.Pow10(p)
	return math.Trunc(f*x+0.5) / x
}

// String2Bytes convert hex-string to []byte
//
//	 args:
//		data: 输入字符串
//		sep： 用于分割字符串的分割字符
//	 return:
//		字节切片
func String2Bytes(data string, sep string) []byte {
	var z []byte
	a := strings.Split(data, sep)
	z = make([]byte, len(a))
	for k, v := range a {
		b, _ := strconv.ParseUint(v, 16, 8)
		z[k] = byte(b)
	}
	return z
}

// Bytes2String convert []byte to hex-string
//
//	 args:
//		data: 输入字节切片
//		sep： 用于分割字符串的分割字符
//	 return:
//		字符串
func Bytes2String(data []byte, sep string) string {
	a := make([]string, len(data))
	for k, v := range data {
		a[k] = fmt.Sprintf("%02x", v)
	}
	return strings.Join(a, sep)
}

// String2Int convert string 2 int
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		int
func String2Int(s string, t int) int {
	x, _ := strconv.ParseInt(s, t, 32)
	return int(x)
}

// String2Byte convert string 2 one byte
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		byte
func String2Byte(s string, t int) byte {
	x, _ := strconv.ParseUint(s, t, 8)
	return byte(x)
}

// String2Int8 convert string 2 int8
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		int8
func String2Int8(s string, t int) byte {
	x, _ := strconv.ParseInt(s, t, 0)
	return byte(x)
}

// String2Int32 convert string 2 int32
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		int32
func String2Int32(s string, t int) int32 {
	x, _ := strconv.ParseInt(s, t, 32)
	return int32(x)
}

// String2Int64 convert string 2 int64
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		int64
func String2Int64(s string, t int) int64 {
	x, _ := strconv.ParseInt(s, t, 64)
	return x
}

// String2UInt64 convert string 2 uint64
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		uint64
func String2UInt64(s string, t int) uint64 {
	x, _ := strconv.ParseUint(s, t, 64)
	return x
}

// String2Float32 convert string 2 float64
func String2Float32(s string) float32 {
	x, _ := strconv.ParseFloat(s, 32)
	return float32(x)
}

// String2Float64 convert string 2 float64
func String2Float64(s string) float64 {
	x, _ := strconv.ParseFloat(s, 64)
	return x
}

// StringSlice2Int8 convert string Slice 2 int8
func StringSlice2Int8(bs []string) byte {
	return String2Byte(strings.Join(bs, ""), 2)
}

// Stamp2Time convert stamp to datetime string
func Stamp2Time(t int64, fmt ...string) string {
	var f string
	if len(fmt) > 0 {
		f = fmt[0]
	} else {
		f = "2006-01-02 15:04:05"
	}
	tm := time.Unix(t, 0)
	return tm.Format(f)
}

// Time2Stampf 可根据制定的时间格式和时区转换为当前时区的Unix时间戳
//
//		s：时间字符串
//	 fmt：时间格式
//	 year：2006，month：01，day：02
//	 hour：15，minute：04，second：05
//	 tz：0～12,超范围时使用本地时区
func Time2Stampf(s, fmt string, tz float32) int64 {
	if s == "" {
		return 0
	}
	if fmt == "" {
		fmt = DateTimeFormat
	}
	if tz > 12 || tz < 0 {
		_, t := time.Now().Zone()
		tz = float32(t / 3600)
	}
	loc := time.FixedZone("", int((time.Duration(tz) * time.Hour).Seconds()))
	tm, ex := time.ParseInLocation(fmt, s, loc)
	if ex != nil {
		return 0
	}
	return tm.Unix()
}

// Time2Stamp convert datetime string to stamp
func Time2Stamp(s string) int64 {
	return Time2Stampf(s, "", 8)
}

// Time2StampNB 电信NB平台数据时间戳转为本地unix时间戳
func Time2StampNB(s string) int64 {
	return Time2Stampf(s, "20060102T150405Z", 0)
}

// SwitchStamp switch stamp format between unix and c#
func SwitchStamp(t int64) int64 {
	y := int64(621356256000000000)
	z := int64(10000000)
	if t > y {
		return (t - y) / z
	}
	return t*z + y
}

// Byte2Bytes int8 to bytes
func Byte2Bytes(v byte, reverse bool) []byte {
	s := fmt.Sprintf("%08b", v)
	if reverse {
		s = ReverseString(s)
	}
	b := make([]byte, 0)
	for _, v := range s {
		if v == 48 {
			b = append(b, 0)
		} else {
			b = append(b, 1)
		}
	}
	return b
}

// Byte2Int32s int8 to int32 list
func Byte2Int32s(v byte, reverse bool) []int32 {
	s := fmt.Sprintf("%08b", v)
	if reverse {
		s = ReverseString(s)
	}
	b := make([]int32, 0)
	for _, v := range s {
		if v == 48 {
			b = append(b, 0)
		} else {
			b = append(b, 1)
		}
	}
	return b
}

// Bcd2Int8 bcd to int
func Bcd2Int8(v byte) byte {
	return ((v&0xf0)>>4)*10 + (v & 0x0f)
}

// Int82Bcd int to bcd
func Int82Bcd(v byte) byte {
	return ((v / 10) << 4) | (v % 10)
}

// Uint642Bytes 长整形转换字节数组（8位），bigOrder==true，高位在前
func Uint642Bytes(i uint64, bigOrder bool) []byte {
	buf := make([]byte, 8)
	if bigOrder {
		binary.BigEndian.PutUint64(buf, i)
	} else {
		binary.LittleEndian.PutUint64(buf, i)
	}
	return buf
}

// Int642Bytes 无符号长整形转换字节数组（8位），bigOrder==true，高位在前
func Int642Bytes(i int64, bigOrder bool) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if bigOrder {
		binary.Write(bytesBuffer, binary.BigEndian, &i)
	} else {
		binary.Write(bytesBuffer, binary.LittleEndian, &i)
	}
	return bytesBuffer.Bytes()
}

// Bytes2Float64 字节数组转双精度浮点，bigOrder==true,高位在前
func Bytes2Float64(b []byte, bigOrder bool) float64 {
	if len(b) < 8 {
		return 0
	}
	if bigOrder {
		return math.Float64frombits(binary.BigEndian.Uint64(b))
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
	// return math.Float64frombits(Bytes2Uint64(b, bigOrder))
}

// Bytes2Float32 字节数组转单精度浮点，bigOrder==true,高位在前
func Bytes2Float32(b []byte, bigOrder bool) float32 {
	if len(b) < 4 {
		return 0
	}
	if bigOrder {
		return math.Float32frombits(binary.BigEndian.Uint32(b))
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
	// return math.Float32frombits(uint32(Bytes2Uint64(b, bigOrder)))
}

// Imgfile2Base64 图片转base64
func Imgfile2Base64(s string) (string, error) {
	f, err := os.ReadFile(s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(f), nil
}

// Base64Imgfile base64码保存为图片
func Base64Imgfile(b, f string) error {
	a, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		return err
	}
	return os.WriteFile(f, a, 0o666)
}

// SplitStringWithLen 按制定长度分割字符串
//
//	s-原始字符串
//	l-切割长度
func SplitStringWithLen(s string, l int) []string {
	rs := []rune(s)
	ss := make([]string, 0)
	xs := ""
	for k, v := range rs {
		xs = xs + string(v)
		if (k+1)%l == 0 {
			ss = append(ss, xs)
			xs = ""
		}
	}
	if len(xs) > 0 {
		ss = append(ss, xs)
	}
	return ss
}

// HexString2Bytes 转换hexstring为字节数组
//
//	s-hexstring（11aabb）
//	bigorder-是否高位在前
//	false低位在前
func HexString2Bytes(s string, bigorder bool) []byte {
	if len(s)%2 == 1 {
		s = "0" + s
	}
	ss := SplitStringWithLen(s, 2)
	b := make([]byte, len(ss))
	if bigorder {
		for k, v := range ss {
			b[k] = String2Byte(v, 16)
		}
	} else {
		c := 0
		for i := len(ss) - 1; i >= 0; i-- {
			b[c] = String2Byte(ss[i], 16)
			c++
		}
	}
	return b
}

// Bytes2Uint64 字节数组转换为uint64
//
//	b-字节数组
//	bigorder-是否高位在前
//	false低位在前
func Bytes2Uint64(b []byte, bigorder bool) uint64 {
	s := ""
	for _, v := range b {
		if bigorder {
			s = s + fmt.Sprintf("%02x", v)
		} else {
			s = fmt.Sprintf("%02x", v) + s
		}
	}
	u, _ := strconv.ParseUint(s, 16, 64)
	return u
}

// Bytes2Int64 字节数组转换为int64
//
//	b-字节数组
//	bigorder-是否高位在前
//	false低位在前
func Bytes2Int64(b []byte, bigorder bool) int64 {
	s := ""
	for _, v := range b {
		if bigorder {
			s = s + fmt.Sprintf("%02x", v)
		} else {
			s = fmt.Sprintf("%02x", v) + s
		}
	}
	u, _ := strconv.ParseInt(s, 16, 64)
	return u
}

// GPS2DFM 经纬度转度分秒
func GPS2DFM(l float64) (int, int, float64) {
	du, x := math.Modf(l)
	fen, y := math.Modf(x * 60)
	return int(du), int(fen), y * 60
}

// DFM2GPS 度分秒转经纬度
func DFM2GPS(du, fen int, miao float64) float64 {
	return float64(du) + float64(fen)/60 + miao/360000
}

// Float642BcdBytes 浮点转bcd字节数组（小端序）
//
//	v：十进制浮点数值
//	f：格式化的字符串，如"%07.03f","%03.0f"
func Float642BcdBytes(v float64, f string) []byte {
	s := strings.ReplaceAll(fmt.Sprintf(f, math.Abs(v)), ".", "")
	var b bytes.Buffer
	if len(s)%2 != 0 {
		s = "0" + s
	}
	for i := len(s); i > 1; i -= 2 {
		if i == 2 {
			if v >= 0 {
				b.WriteByte(String2Byte(s[i-2:i], 16))
			} else {
				b.WriteByte(String2Byte(s[i-2:i], 16) + 0x80)
			}
		} else {
			b.WriteByte(String2Byte(s[i-2:i], 16))
		}
	}
	return b.Bytes()
}

// BcdBytes2Float64 bcd数组转浮点(小端序)
//
//	b:bcd数组
//	d：小数位数
//	Unsigned：无符号的
func BcdBytes2Float64(b []byte, decimal int, unsigned bool) float64 {
	negative := false
	var s string
	for k, v := range b {
		if k == len(b)-1 { // 最后一位，判正负
			if !unsigned {
				if v >= 128 {
					v = v - 0x80
					negative = true
				}
			}
		}
		s = fmt.Sprintf("%02x", v) + s
	}
	s = s[:len(s)-decimal] + "." + s[len(s)-decimal:]
	f, _ := strconv.ParseFloat(s, 64)
	if negative {
		f = f * -1
	}
	return f
}

// Float642BcdBytesBigOrder 浮点转bcd字节数组（大端序）
//
//	v：十进制浮点数值
//	f：格式化的字符串，如"%07.03f","%03.0f"
func Float642BcdBytesBigOrder(v float64, f string) []byte {
	s := strings.ReplaceAll(fmt.Sprintf(f, math.Abs(v)), ".", "")
	var b bytes.Buffer
	if len(s)%2 != 0 {
		s = "0" + s
	}
	for i := 0; i < len(s); i += 2 {
		if i == 2 {
			if v > 0 {
				b.WriteByte(String2Byte(s[i:i+2], 16))
			} else {
				b.WriteByte(String2Byte(s[i:i+2], 16) + 0x80)
			}
		} else {
			b.WriteByte(String2Byte(s[i:i+2], 16))
		}
	}
	return b.Bytes()
}

// BcdBytes2Float64BigOrder bcd数组转浮点(大端序)
//
//	b:bcd数组
//	d：小数位数
//	Unsigned：无符号的
func BcdBytes2Float64BigOrder(b []byte, decimal int, unsigned bool) float64 {
	negative := false
	var s string
	for k, v := range b {
		if k == len(b)-1 { // 最后一位，判正负
			if !unsigned {
				if v >= 128 {
					v = v - 0x80
					negative = true
				}
			}
		}
		s += fmt.Sprintf("%02x", v)
	}
	s = s[:len(s)-decimal] + "." + s[len(s)-decimal:]
	f, _ := strconv.ParseFloat(s, 64)
	if negative {
		f = f * -1
	}
	return f
}

// Bcd2STime bcd转hh*60+mm
func Bcd2STime(b []byte) int32 {
	return String2Int32(fmt.Sprintf("%02x", b[0]), 10)*60 + String2Int32(fmt.Sprintf("%02x", b[1]), 10)
}

// STime2Bcd hh*60+mm转BCD
func STime2Bcd(t int32) []byte {
	// return []byte{Bcd2Int8(byte(t / 60)), Bcd2Int8(byte(t % 60))}
	return []byte{Int82Bcd(byte(t / 60)), Int82Bcd(byte(t % 60))}
}

// SignedInt322Byte 有符号整形转byte
//
// Deprecated: use Int82Byte()
func SignedInt322Byte(i int32) byte {
	return Int82Byte(int8(i))
}

// Byte2SignedInt32 byte转有符号整型
//
// Deprecated: use Byte2Int8()
func Byte2SignedInt32(b byte) int32 {
	return int32(Byte2Int8(b))
}

// Int82Byte 有符号整型转byte
func Int82Byte(i int8) byte {
	return uint8(i)
}

// Byte2Int8 byte转有符号整型
func Byte2Int8(b byte) int8 {
	return int8(b)
	// if b <= 127 {
	// 	return int32(b)
	// }
	// return 0 - (int32(^b<<1>>1) + 1)
}

// BcdDT2Stamp bcd时间戳转unix
func BcdDT2Stamp(d []byte) int64 {
	f := "0601021504"
	if len(d) == 6 {
		f = "060102150405"
	}
	return Time2Stampf(strconv.FormatInt(int64(BcdBytes2Float64(d, 0, true)), 10), f, 8)
}

// Stamp2BcdDT unix时间戳转bcd,6字节，第一字节为秒
func Stamp2BcdDT(dt int64) []byte {
	if dt == 0 {
		return []byte{0, 0, 0, 0, 0, 0}
	}
	return Float642BcdBytes(String2Float64(Stamp2Time(dt, "060102150405")), "%12.0f")
}

// EncodeUTF16BE 将字符串编码成utf16be的格式，用于cdma短信发送
func EncodeUTF16BE(s string) []byte {
	a := utf16.Encode([]rune(s))
	var b bytes.Buffer
	for _, v := range a {
		b.Write([]byte{byte(v >> 8), byte(v)})
	}
	return b.Bytes()
}

// String2Unicode 字符串转4位unicode编码
func String2Unicode(s string) string {
	var str string
	for _, v := range s {
		str += fmt.Sprintf("%04X", v)
	}
	return str
}

// SMSUnicode 编码短信
func SMSUnicode(s string) []string {
	return SplitStringWithLen(String2Unicode(s), 67*4)
}

// SignedBytes2Int64 有符号字节数组转int64,低前
func SignedBytes2Int64(b []byte) int64 {
	s := ""
	for _, v := range b {
		s = fmt.Sprintf("%08b", v) + s
	}
	x, err := strconv.ParseInt(s[1:], 2, 64)
	if err != nil {
		return 0
	}
	switch s[0] {
	case 49:
		return -1 * x
	default:
		return x
	}
}

// Days2String 将天数转换为年月日显示
func Days2String(days int) string {
	t1, _ := time.Parse("2006-01-02", "0000-00-00")
	t2 := t1.Add(time.Hour * time.Duration(24*days))
	y := t2.Year() - 1
	if y < 0 {
		y = 0
	}
	m := int(t2.Month()) - 1
	if m < 0 {
		m = 0
	}
	d := t2.Day() - 1
	if d < 0 {
		d = 0
	}
	out := []string{}
	if y > 0 {
		out = append(out, fmt.Sprintf("%d Years", y))
	}
	if m > 0 {
		out = append(out, fmt.Sprintf("%d Months", m))
	}
	if d == 0 {
		out = append(out, "less than a day")
	} else {
		if d > 0 {
			out = append(out, fmt.Sprintf("%d Days", d))
		}
	}
	return strings.Join(out, ", ")
}

// Seconds2String 秒数转换成天，小时，分钟
func Seconds2String(sec int64) string {
	var days, hours, minutes int64
	days = sec / 60 / 60 / 24
	if a := sec - days*60*60*24; a > 0 {
		hours = a / 60 / 60
	}
	if a := sec - days*60*60*24 - hours*60*60; a > 0 {
		minutes = a / 60
	}
	out := []string{}
	if days > 0 {
		out = append(out, fmt.Sprintf("%d Days", days))
	}
	if hours > 0 {
		out = append(out, fmt.Sprintf("%d Hours", hours))
	}
	if minutes+hours+days == 0 {
		out = append(out, "less than a minute")
	} else {
		if minutes > 0 {
			out = append(out, fmt.Sprintf("%d Minutes", minutes))
		}
	}
	return strings.Join(out, ", ")
}

// Days2StringCHS 将天数转换为年月日显示
func Days2StringCHS(days int) string {
	t1, _ := time.Parse("2006-01-02", "0000-00-00")
	t2 := t1.Add(time.Hour * time.Duration(24*days))
	y := t2.Year() - 1
	if y < 0 {
		y = 0
	}
	m := int(t2.Month()) - 1
	if m < 0 {
		m = 0
	}
	d := t2.Day() - 1
	if d < 0 {
		d = 0
	}
	out := []string{}
	if y > 0 {
		out = append(out, fmt.Sprintf("%d年", y))
	}
	if m > 0 {
		out = append(out, fmt.Sprintf("%d个月", m))
	}
	if m+y+d == 0 {
		out = append(out, "不到1天")
	} else {
		if d > 0 {
			out = append(out, fmt.Sprintf("%d天", d))
		}
	}
	return strings.Join(out, ", ")
}

// Seconds2StringCHS 秒数转换成天，小时，分钟
func Seconds2StringCHS(sec int64) string {
	var days, hours, minutes int64
	days = sec / 60 / 60 / 24
	if a := sec - days*60*60*24; a > 0 {
		hours = a / 60 / 60
	}
	if a := sec - days*60*60*24 - hours*60*60; a > 0 {
		minutes = a / 60
	}
	out := []string{}
	if days > 0 {
		out = append(out, fmt.Sprintf("%d天", days))
	}
	if hours > 0 {
		out = append(out, fmt.Sprintf("%d小时", hours))
	}
	if minutes+hours+days == 0 {
		out = append(out, "不到一分钟")
	} else {
		if minutes > 0 {
			out = append(out, fmt.Sprintf("%d分钟", minutes))
		}
	}
	return strings.Join(out, ", ")
}
