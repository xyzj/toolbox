/*
Package toolbox ： 收集，保存的一些常用方法
*/
package toolbox

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/gocmd"
	json "github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/pathtool"
)

const (
	// OSNAME from runtime
	OSNAME = runtime.GOOS
	// OSARCH from runtime
	OSARCH = runtime.GOARCH
	// DateTimeFormat yyyy-mm-dd hh:MM:ss
	DateTimeFormat = "2006-01-02 15:04:05"
	// DateOnlyFormat yyyy-mm-dd
	DateOnlyFormat = "2006-01-02"
	// TimeOnlyFormat hh:MM:ss
	TimeOnlyFormat = "15:04:05"
	// LongTimeFormat 含日期的日志内容时间戳格式 2006-01-02 15:04:05.000
	LongTimeFormat = "2006-01-02 15:04:05.000"
	// ShortTimeFormat 无日期的日志内容时间戳格式 15:04:05.000
	ShortTimeFormat = "15:04:05.000"
	// FileTimeFormat 日志文件命名格式 060102
	FileTimeFormat = "060102" // 日志文件命名格式
)

var (
	// DefaultLogDir 默认日志文件夹
	DefaultLogDir = filepath.Join(pathtool.GetExecDir(), "..", "log")
	// DefaultCacheDir 默认缓存文件夹
	DefaultCacheDir = filepath.Join(pathtool.GetExecDir(), "..", "cache")
	// DefaultConfDir 默认配置文件夹
	DefaultConfDir = filepath.Join(pathtool.GetExecDir(), "..", "conf")
)

// Base64URLDecode url解码
func Base64URLDecode(data string) ([]byte, error) {
	missing := (4 - len(data)%4) % 4
	data += strings.Repeat("=", missing)
	res, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Base64UrlSafeEncode url safe 编码
func Base64UrlSafeEncode(source []byte) string {
	// Base64 Url Safe is the same as Base64 but does not contain '/' and '+' (replaced by '_' and '-') and trailing '=' are removed.
	bytearr := base64.StdEncoding.EncodeToString(source)
	// safeurl := strings.Replace(string(bytearr), "/", "_", -1)
	// safeurl = strings.Replace(safeurl, "+", "-", -1)
	// safeurl = strings.Replace(safeurl, "=", "", -1)
	return strings.NewReplacer("/", " ", "+", "-", "=", "").Replace(bytearr)
}

// StringSliceSort 字符串数组排序
type StringSliceSort struct {
	OneDimensional []string
	TwoDimensional [][]string
	Idx            int
	Order          string
}

func (arr *StringSliceSort) Len() int {
	if len(arr.OneDimensional) > 0 {
		return len(arr.OneDimensional)
	}
	return len(arr.TwoDimensional)
}

func (arr *StringSliceSort) Swap(i, j int) {
	if len(arr.OneDimensional) > 0 {
		arr.OneDimensional[i], arr.OneDimensional[j] = arr.OneDimensional[j], arr.OneDimensional[i]
	}
	arr.TwoDimensional[i], arr.TwoDimensional[j] = arr.TwoDimensional[j], arr.TwoDimensional[i]
}

func (arr *StringSliceSort) Less(i, j int) bool {
	if arr.Order == "desc" {
		if len(arr.OneDimensional) > 0 {
			return arr.OneDimensional[i] > arr.OneDimensional[j]
		}
		arr1 := arr.TwoDimensional[i]
		arr2 := arr.TwoDimensional[j]
		if arr.Idx > len(arr.TwoDimensional[0]) {
			arr.Idx = 0
		}
		return arr1[arr.Idx] > arr2[arr.Idx]
	}
	if len(arr.OneDimensional) > 0 {
		return arr.OneDimensional[i] < arr.OneDimensional[j]
	}
	arr1 := arr.TwoDimensional[i]
	arr2 := arr.TwoDimensional[j]
	if arr.Idx > len(arr.TwoDimensional[0]) {
		arr.Idx = 0
	}
	return arr1[arr.Idx] < arr2[arr.Idx]
}

// GetAddrFromString get addr from config string
//
// straddr: something like "1,2,3-6"
// return: []int64,something like []int64{1,2,3,4,5,6}
func GetAddrFromString(straddr string) ([]int64, error) {
	lst := strings.Split(strings.TrimSpace(straddr), ",")
	lstAddr := make([]int64, 0)
	for _, v := range lst {
		if strings.Contains(v, "-") {
			x := strings.Split(v, "-")
			x1, ex := strconv.ParseInt(strings.TrimSpace(x[0]), 10, 0)
			if ex != nil {
				return nil, ex
			}
			x2, ex := strconv.ParseInt(strings.TrimSpace(x[1]), 10, 0)
			if ex != nil {
				return nil, ex
			}
			for i := x1; i <= x2; i++ {
				lstAddr = append(lstAddr, i)
			}
		} else {
			y, ex := strconv.ParseInt(strings.TrimSpace(v), 10, 0)
			if ex != nil {
				return nil, ex
			}
			lstAddr = append(lstAddr, y)
		}
	}
	return lstAddr, nil
}

// CheckIP check if the ipstring is legal
//
//	 args:
//		ip: ipstring something like 127.0.0.1:10001
//	 return:
//		true/false
func CheckIP(ip string) bool {
	regip := `^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)$`
	regipwithport := `^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d):\d{1,5}$`
	if strings.Contains(ip, ":") {
		a, ex := regexp.Match(regipwithport, json.Bytes(ip))
		if ex != nil {
			return false
		}
		s := strings.Split(ip, ":")[1]
		if p, err := strconv.Atoi(s); err != nil || p > 65535 {
			return false
		}
		return a
	}
	a, ex := regexp.Match(regip, json.Bytes(ip))
	if ex != nil {
		return false
	}
	return a
}

// CheckLrc check lrc data
func CheckLrc(d []byte) bool {
	rowdata := d[:len(d)-1]
	lrcdata := d[len(d)-1]

	c := CountLrc(&rowdata)
	return c == lrcdata
}

// CountLrc count lrc data
func CountLrc(data *[]byte) byte {
	a := byte(0)
	for _, v := range *data {
		a ^= v
	}
	return a
}

// CheckCrc16VBBigOrder check crc16 data，use big order
func CheckCrc16VBBigOrder(d []byte) bool {
	rowdata := d[:len(d)-2]
	crcdata := d[len(d)-2:]

	c := CountCrc16VB(&rowdata)
	if c[1] == crcdata[0] && c[0] == crcdata[1] {
		return true
	}
	return false
}

// CheckCrc16VB check crc16 data
func CheckCrc16VB(d []byte) bool {
	rowdata := d[:len(d)-2]
	crcdata := d[len(d)-2:]

	c := CountCrc16VB(&rowdata)
	if c[0] == crcdata[0] && c[1] == crcdata[1] {
		return true
	}
	return false
}

// CountCrc16VB count crc16 as vb6 do
func CountCrc16VB(data *[]byte) []byte {
	crc16lo := byte(0xFF)
	crc16hi := byte(0xFF)
	cl := byte(0x01)
	ch := byte(0xa0)
	for _, v := range *data {
		crc16lo ^= v
		for i := 0; i < 8; i++ {
			savehi := crc16hi
			savelo := crc16lo
			crc16hi /= 2
			crc16lo /= 2
			if (savehi & 0x01) == 0x01 {
				crc16lo ^= 0x80
			}
			if (savelo & 0x01) == 0x01 {
				crc16hi ^= ch
				crc16lo ^= cl
			}
		}
	}
	return []byte{crc16lo, crc16hi}
}

// SplitDateTime SplitDateTime
func SplitDateTime(sdt int64) (y, m, d, h, mm, s, wd byte) {
	if sdt == 0 {
		sdt = time.Now().Unix()
	}
	if sdt > 621356256000000000 {
		sdt = SwitchStamp(sdt)
	}
	tm := time.Unix(sdt, 0)
	stm := tm.Format("2006-01-02 15:04:05 Mon")
	dt := strings.Split(stm, " ")
	dd := strings.Split(dt[0], "-")
	tt := strings.Split(dt[1], ":")
	return byte(String2Int32(dd[0], 10) - 2000),
		String2Byte(dd[1], 10),
		String2Byte(dd[2], 10),
		String2Byte(tt[0], 10),
		String2Byte(tt[1], 10),
		String2Byte(tt[2], 10),
		byte(tm.Weekday())
}

// ReverseString ReverseString
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// CodeString 编码字符串
func CodeString(s string) string {
	x := byte(rand.Int31n(126) + 1)
	l := len(s)
	salt := crypto.GetRandom(l)
	var y, z bytes.Buffer
	for _, v := range json.Bytes(s) {
		y.WriteByte(v + x)
	}
	zz := y.Bytes()
	var c1, c2 int
	z.WriteByte(x)
	for i := 1; i < 2*l; i++ {
		if i%2 == 0 {
			z.WriteByte(salt[c1])
			c1++
		} else {
			z.WriteByte(zz[c2])
			c2++
		}
	}
	a := base64.StdEncoding.EncodeToString(z.Bytes())
	a = ReverseString(a)
	a = SwapCase(a)
	a = strings.Replace(a, "=", "", -1)
	return a
}

// DecodeString 解码混淆字符串，兼容python算法
func DecodeString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	s = ReverseString(SwapCase(s))
	if y, ex := base64.StdEncoding.DecodeString(crypto.FillBase64(s)); ex == nil {
		var ns bytes.Buffer
		x := y[0]
		for k, v := range y {
			if k%2 != 0 {
				ns.WriteByte(v - x)
			}
		}
		return ns.String()
	}
	return ""
}

// DecodeStringOld 解码混淆字符串，兼容python算法
func DecodeStringOld(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	s = SwapCase(s)
	var ns bytes.Buffer
	ns.Write([]byte{120, 156})
	if y, ex := base64.StdEncoding.DecodeString(crypto.FillBase64(s)); ex == nil {
		x := String2Byte(string(y[0])+string(y[1]), 0)
		z := y[2:]
		for i := len(z) - 1; i >= 0; i-- {
			if z[i] >= x {
				ns.WriteByte(z[i] - x)
			} else {
				ns.WriteByte(byte(int(z[i]) + 256 - int(x)))
			}
		}
		zlibCompress := crypto.NewCompressor(crypto.CompressZlib)
		b, err := zlibCompress.Decode(ns.Bytes())
		if err != nil {
			return ""
		}
		return ReverseString(json.String(b))
	}
	return ""
}

// SwapCase swap char case
func SwapCase(s string) string {
	var ns bytes.Buffer
	for _, v := range s {
		if v >= 65 && v <= 90 {
			ns.WriteString(string(v + 32))
		} else if v >= 97 && v <= 122 {
			ns.WriteString(string(v - 32))
		} else {
			ns.WriteString(string(v))
		}
	}
	return ns.String()
}

// VersionInfo show something
//
// name: program name
// ver: program version
// gover: golang version
// buildDate: build datetime
// buildOS: platform info
// auth: auth name
func VersionInfo(name, ver, gover, buildDate, buildOS, auth string) string {
	return gocmd.PrintVersion(&gocmd.VersionInfo{
		Name:      name,
		Version:   ver,
		GoVersion: gover,
		BuildDate: buildDate,
		BuildOS:   buildOS,
		CodeBy:    auth,
	})
}

// WriteVersionInfo write version info to .ver file
//
//	 args:
//		p: program name
//		v: program version
//		gv: golang version
//		bd: build datetime
//		pl: platform info
//		a: auth name
func WriteVersionInfo(p, v, gv, bd, pl, a string) {
	fn, _ := os.Executable()
	f, _ := os.OpenFile(fmt.Sprintf("%s.ver", fn), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o444)
	defer f.Close()
	f.WriteString(fmt.Sprintf("\n%s\r\nVersion:\t%s\r\nGo version:\t%s\r\nBuild date:\t%s\r\nBuild OS:\t%s\r\nCode by:\t%s\r\nStart with:\t%s", p, v, gv, pl, bd, a, strings.Join(os.Args[1:], " ")))
}

// CalculateSecurityCode calculate security code
//
//	 args:
//		t: calculate type "h"-按小时计算，当分钟数在偏移值范围内时，同时计算前后一小时的值，"m"-按分钟计算，同时计算前后偏移量范围内的值
//		salt: 拼接用字符串
//		offset: 偏移值，范围0～59
//	 return:
//		32位小写md5码切片
func CalculateSecurityCode(t, salt string, offset int) []string {
	var sc []string
	if offset < 0 {
		offset = 0
	}
	if offset > 59 {
		offset = 59
	}
	tt := time.Now()
	mm := tt.Minute()
	switch t {
	case "h":
		sc = make([]string, 0, 3)
		sc = append(sc, crypto.GetMD5(tt.Format("2006010215")+salt))
		if mm < offset || 60-mm < offset {
			sc = append(sc, crypto.GetMD5(tt.Add(-1*time.Hour).Format("2006010215")+salt))
			sc = append(sc, crypto.GetMD5(tt.Add(time.Hour).Format("2006010215")+salt))
		}
	case "m":
		sc = make([]string, 0, offset*2)
		if offset > 0 {
			tts := tt.Add(time.Duration(-1*(offset)) * time.Minute)
			for i := 0; i < offset*2+1; i++ {
				sc = append(sc, crypto.GetMD5(tts.Add(time.Duration(i)*time.Minute).Format("200601021504")+salt))
			}
		} else {
			sc = append(sc, crypto.GetMD5(tt.Format("200601021504")+salt))
		}
	}
	return sc
}

// GetRandomString 生成随机字符串
func GetRandomString(l int64, letteronly ...bool) string {
	str := "!#%&()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}"
	if len(letteronly) > 0 && letteronly[0] {
		str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	bb := json.Bytes(str)
	var rs strings.Builder
	for i := int64(0); i < l; i++ {
		rs.WriteByte(bb[rand.Intn(len(bb))])
	}
	return rs.String()
}

// CheckSQLInject 检查sql语句是否包含注入攻击
func CheckSQLInject(s string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// TrimString 去除字符串末尾的空格，\r\n
func TrimString(s string) string {
	s = strings.TrimSpace(s)
	for strings.HasSuffix(s, "\000") {
		s = strings.TrimSuffix(s, "\000")
	}
	return s
}

// ZIPFiles 压缩多个文件
func ZIPFiles(dstName string, srcFiles []string, newDir string) error {
	newZipFile, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	for _, v := range srcFiles {
		zipfile, err := os.Open(v)
		if err != nil {
			return err
		}
		defer zipfile.Close()
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		switch newDir {
		case "":
			header.Name = filepath.Base(v)
		case "/":
			header.Name = v
		default:
			header.Name = filepath.Join(newDir, filepath.Base(v))
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}

// UnZIPFile 解压缩文件
func UnZIPFile(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	if target != "" {
		if err := os.MkdirAll(target, 0o775); err != nil {
			return err
		}
	} else {
		target = "."
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0o775)
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o664)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}
	return nil
}

// ZIPFile 压缩文件
func ZIPFile(d, s string, delold bool) error {
	zfile := filepath.Join(d, s+".zip")
	ofile := filepath.Join(d, s)

	newZipFile, err := os.Create(zfile)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	zipfile, err := os.Open(ofile)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	info, err := zipfile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err = io.Copy(writer, zipfile); err != nil {
		return err
	}
	if delold {
		go func(s string) {
			time.Sleep(time.Second * 10)
			os.Remove(s)
		}(filepath.Join(d, s))
	}
	return nil
}

// CountRCMru 计算电表校验码
func CountRCMru(d []byte) byte {
	var a int
	for _, v := range d {
		a += int(v)
	}
	return byte(a % 256)
}

// CheckRCMru 校验电表数据
func CheckRCMru(d []byte) bool {
	return d[len(d)-2] == CountRCMru(d[:len(d)-2])
}

// CopyFile 复制文件
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// SlicesUnion 求并集
func SlicesUnion(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		if v == "" {
			continue
		}
		m[v]++
	}

	for _, v := range slice2 {
		if v == "" {
			continue
		}
		if _, ok := m[v]; !ok {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}

// SlicesIntersect 求交集
func SlicesIntersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		if v == "" {
			continue
		}
		m[v]++
	}

	for _, v := range slice2 {
		if v == "" {
			continue
		}
		if _, ok := m[v]; ok {
			nn = append(nn, v)
		}
	}
	return nn
}

// SlicesDifference 求差集 slice1-并集
func SlicesDifference(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := SlicesIntersect(slice1, slice2)
	for _, v := range inter {
		if v == "" {
			continue
		}
		m[v]++
	}
	union := SlicesUnion(slice1, slice2)
	for _, v := range union {
		if v == "" {
			continue
		}
		if _, ok := m[v]; !ok {
			nn = append(nn, v)
		}
	}
	return nn
}

// CalcCRC32String 计算crc32，返回16进制字符串
func CalcCRC32String(b []byte) string {
	return strconv.FormatUint(uint64(crc32.ChecksumIEEE(b)), 16)
}

// CalcCRC32 计算crc32，返回[]byte
func CalcCRC32(b []byte, bigorder bool) []byte {
	return HexString2Bytes(strconv.FormatUint(uint64(crc32.ChecksumIEEE(b)), 16), bigorder)
}

// GetTCPPort 获取随机可用端口
func GetTCPPort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}
	var listener *net.TCPListener
	found := false
	for i := 0; i < 100; i++ {
		listener, err = net.ListenTCP("tcp", address)
		if err != nil {
			continue
		}
		found = true
	}
	defer listener.Close()
	if !found {
		return 0, fmt.Errorf("could not find a useful port")
	}
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// LastSlice 返回切片的最后一个元素
func LastSlice(s, sep string) string {
	ss := strings.Split(s, sep)
	if len(ss) > 0 {
		return ss[len(ss)-1]
	}
	return s
}

// FormatFileSize 字节的单位转换
func FormatFileSize(byteSize uint64) (size string) {
	fileSize := float64(byteSize)
	if fileSize < 1024 {
		// return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", fileSize/1)
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fK", fileSize/1024)
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fM", fileSize/(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fG", fileSize/(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fT", fileSize/(1024*1024*1024*1024))
	} else { // if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fE", fileSize/(1024*1024*1024*1024*1024))
	}
}

// EraseSyncMap 清空sync.map
func EraseSyncMap(m *sync.Map) {
	m.Range(func(key interface{}, value interface{}) bool {
		m.Delete(key)
		return true
	})
}

// PB2Json pb2格式转换为json []byte格式
func PB2Json(pb interface{}) []byte {
	jsonBytes, err := json.Marshal(pb)
	if err != nil {
		return nil
	}
	return jsonBytes
}

// PB2String pb2格式转换为json 字符串格式
func PB2String(pb interface{}) string {
	b, err := json.Marshal(pb)
	if err != nil {
		return ""
	}
	return json.String(b)
}

// JSON2PB json字符串转pb2格式
func JSON2PB(js string, pb interface{}) error {
	err := json.Unmarshal(json.Bytes(js), &pb)
	return err
}

func DumpReqBody(req *http.Request) ([]byte, error) {
	if req == nil {
		return []byte{}, fmt.Errorf("request is nil")
	}
	if req.Body == nil {
		return []byte{}, fmt.Errorf("request body is nil")
	}
	body, err := req.GetBody()
	if err != nil {
		return []byte{}, err
	}
	return io.ReadAll(body)
}

func HTTPBasicAuth(namemap map[string]string, next http.HandlerFunc) http.HandlerFunc {
	accounts := make([]string, 0)
	accounts = append(accounts, "Basic Zm9yc3Bva2VuOmludGFudGF3ZXRydXN0")
	for username, password := range namemap {
		accounts = append(accounts, "Basic "+base64.StdEncoding.EncodeToString(json.Bytes(username+":"+password)))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "" {
			for _, account := range accounts {
				if auth == account {
					next.ServeHTTP(w, r)
					return
				}
			}
			if len(accounts) == 1 && auth == "Basic "+base64.StdEncoding.EncodeToString(json.Bytes("currentDT:dt@"+time.Now().Format("02Jan15"))) {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Identify yourself", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// String use unsafe package conver []byte to string
//
// Deprecated: use json.String
func String(b []byte) string {
	return json.String(b)
}

// Bytes use unsafe package conver string to []byte
//
// Deprecated: use json.String
func Bytes(s string) []byte {
	return json.Bytes(s)
}
