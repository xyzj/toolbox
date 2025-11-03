package toolbox

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xyzj/toolbox/httpclient"
)

var httpClient = httpclient.New()

func DoStreamRequest(req *http.Request, header func(map[string]string), recv func([]byte) error, opts ...httpclient.ReqOptions) error {
	return httpClient.DoStreamRequest(req, header, recv, opts...)
}

// DoRequestWithTimeout 发起请求
func DoRequestWithTimeout(req *http.Request, timeo time.Duration) (int, []byte, map[string]string, error) {
	return httpClient.DoRequest(req, httpclient.WithTimeout(timeo))
}

// IPUint2String change ip int64 data to string format
func IPUint2String(ipnr uint) string {
	return fmt.Sprintf("%d.%d.%d.%d", (ipnr>>24)&0xFF, (ipnr>>16)&0xFF, (ipnr>>8)&0xFF, ipnr&0xFF)
}

// IPInt642String change ip int64 data to string format
func IPInt642String(ipnr int64) string {
	return fmt.Sprintf("%d.%d.%d.%d", (ipnr)&0xFF, (ipnr>>8)&0xFF, (ipnr>>16)&0xFF, (ipnr>>24)&0xFF)
}

// IPInt642Bytes change ip int64 data to string format
func IPInt642Bytes(ipnr int64) []byte {
	return []byte{byte((ipnr) & 0xFF), byte((ipnr >> 8) & 0xFF), byte((ipnr >> 16) & 0xFF), byte((ipnr >> 24) & 0xFF)}
}

// IPUint2Bytes change ip int64 data to string format
func IPUint2Bytes(ipnr int64) []byte {
	return []byte{byte((ipnr >> 24) & 0xFF), byte((ipnr >> 16) & 0xFF), byte((ipnr >> 8) & 0xFF), byte((ipnr) & 0xFF)}
}

// IP2Uint change ip string data to int64 format
func IP2Uint(ipnr string) uint {
	// ex := errors.New("wrong ip address format")
	bits := strings.Split(ipnr, ".")
	if len(bits) != 4 {
		return 0
	}
	var intip uint
	for k, v := range bits {
		i, ex := strconv.Atoi(v)
		if ex != nil || i > 255 || i < 0 {
			return 0
		}
		intip += uint(i) << uint(8*(3-k))
	}
	return intip
}

// IP2Int64 change ip string data to int64 format
func IP2Int64(ipnr string) int64 {
	// ex := errors.New("wrong ip address format"
	bits := strings.Split(ipnr, ".")
	if len(bits) != 4 {
		return 0
	}
	var intip uint
	for k, v := range bits {
		i, ex := strconv.Atoi(v)
		if ex != nil || i > 255 || i < 0 {
			return 0
		}
		intip += uint(i) << uint(8*(k))
	}
	return int64(intip)
}

// RealIP 返回本机的v4或v6ip
func RealIP(v6first bool) string {
	s, _ := GetFirstLocalIP(v6first)
	return s
}

// ExternalIP 返回v4地址
// func ExternalIP() string {
// 	v4, v6, err := GlobalIPs()
// 	if err != nil {
// 		return ""
// 	}
// 	if len(v4) > 0 {
// 		return v4[0]
// 	}
// 	if len(v6) > 0 {
// 		return v6[0]
// 	}
// 	return ""
// }

// // ExternalIPV6 返回v6地址
// func ExternalIPV6() string {
// 	_, v6, err := GlobalIPs()
// 	if err != nil {
// 		return ""
// 	}
// 	if len(v6) > 0 {
// 		return v6[0]
// 	}
// 	return ""
// }

// GetLocalIPs 返回本机所有网卡绑定的 IPv4 或 IPv6 地址（不包含回环或未启用接口）。
// 参数 ipv6 = true 则返回 IPv6 地址，否则返回 IPv4 地址。
// 会过滤：未启用接口、loopback、未指定地址（0.0.0.0 / ::）、链路本地地址（如 fe80::/10）。
func GetLocalIPs(ipv6 bool) ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{})
	for _, ifi := range ifaces {
		// 只考虑已启用且非 loopback 的接口
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip == nil {
				continue
			}
			// 排除环回、未指定地址
			if ip.IsLoopback() || ip.IsUnspecified() {
				continue
			}
			if ipv6 {
				// 仅保留 IPv6（排除 IPv4 映射/混合）
				if ip.To4() != nil {
					continue
				}
				// 排除链路本地（fe80::/10）和多播链路本地
				if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
					continue
				}
			} else {
				// 仅保留 IPv4
				if ip4 := ip.To4(); ip4 == nil {
					continue
				}
			}
			s := ip.String()
			// 如果请求 IPv6，返回时加上方括号，方便后续拼接端口如 "[::1]:8080"
			if ipv6 {
				if !strings.HasPrefix(s, "[") && !strings.HasSuffix(s, "]") {
					s = "[" + s + "]"
				}
			}
			set[s] = struct{}{}
		}
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("no local addresses found")
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out, nil
}

// GetFirstLocalIP 返回首个匹配的本机地址，找不到时返回空字符串和错误。
// 优先 IPv4（preferIPv6=false），可通过 preferIPv6 改变优先级。
func GetFirstLocalIP(preferIPv6 bool) (string, error) {
	// 先按偏好获取
	if ips, err := GetLocalIPs(preferIPv6); err == nil && len(ips) > 0 {
		return ips[0], nil
	}
	// 反向再试一次
	if ips, err := GetLocalIPs(!preferIPv6); err == nil && len(ips) > 0 {
		return ips[0], nil
	}
	return "", fmt.Errorf("no local address found")
}

// ValidateIPPort 验证输入的 "ip:port" 字符串（支持 IPv4 和 IPv6）。
// 返回解析出的 ip（不含方括号与 zone）和端口号；解析失败返回 error。
// 支持格式示例：
//
//	1.2.3.4:80
//	[2001:db8::1]:443
//	2001:db8::1:443   // 尝试通过最后一个冒号分割 host/port（推荐使用带方括号的 IPv6）
//
// 注：IPv6 带 zone 的情况会在验证时去掉 zone（例如 fe80::1%eth0 -> fe80::1）。
func ValidateIPPort(s string) (*net.TCPAddr, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	// 优先使用 ResolveTCPAddr（能处理带 zone 的 IPv6、带方括号的 v6 等）
	if a, err := net.ResolveTCPAddr("tcp", s); err == nil {
		if a.Port == 0 {
			return nil, false
		}
		return a, true
	}

	// 回退解析：先尝试 SplitHostPort（处理 [v6]:port）
	var host, portStr string
	if h, p, err := net.SplitHostPort(s); err == nil {
		host = h
		portStr = p
	} else {
		// 使用最后一个冒号分割（处理未带方括号的 v6:port 或 ipv4:port）
		if idx := strings.LastIndex(s, ":"); idx == -1 {
			return nil, false
		} else {
			host = s[:idx]
			portStr = s[idx+1:]
		}
	}

	// 解析端口并检查范围
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil || port <= 0 || port > 65535 {
		return nil, false
	}

	// 去掉方括号并提取 zone（如 fe80::1%eth0）
	host = strings.TrimSpace(host)
	host = strings.Trim(host, "[]")
	zone := ""
	if i := strings.LastIndex(host, "%"); i != -1 {
		zone = host[i+1:]
		host = host[:i]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, false
	}

	tcp := &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	// 如果有 zone 并且是 IPv6，填充 Zone 字段
	if zone != "" && ip.To16() != nil && ip.To4() == nil {
		tcp.Zone = zone
	}

	return tcp, true
}

// Int32SegmentsToIPv6 将 16 字节（每段 2 字节，低字节在前，little-endian per segment）的数据
// 还原为标准的 IPv6 字符串表示（如 "2001:db8::1"）。
// 输入长度必须为 16，否则返回错误。
func Int32SegmentsToIPv6(b []byte) (string, error) {
	if len(b) != 16 {
		return "", fmt.Errorf("invalid input length: %d", len(b))
	}
	ip := make(net.IP, net.IPv6len)
	// 每段 2 字节，原先存储为 [lo, hi], 需要恢复为 [hi, lo]
	for seg := range 8 {
		lo := b[2*seg]
		hi := b[2*seg+1]
		ip[2*seg] = hi
		ip[2*seg+1] = lo
	}
	return ip.String(), nil
}

// IPv6ToInt32Segments 将 IPv6 地址字符串转换为 16 字节的表示形式。
// 每个 16 位段使用低字节在前（little-endian per segment）。
// 支持缩写 ::、带方括号的形式 [::1] 以及带 zone 的形式 fe80::1%eth0。
// 返回长度为 16 的 byte 数组，解析失败返回错误。
func IPv6ToInt32Segments(s string) ([]byte, error) {
	var out = make([]byte, 0, 16)
	s = strings.TrimSpace(s)
	// 去掉方括号 [::1]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}
	// 去掉 zone id 如 %eth0
	if i := strings.LastIndex(s, "%"); i != -1 {
		s = s[:i]
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return out, fmt.Errorf("invalid ip: %s", s)
	}
	ip = ip.To16()
	if ip == nil || ip.To4() != nil {
		return out, fmt.Errorf("not an IPv6 address: %s", s)
	}
	// 每段 2 字节，低字节在前
	for seg := 0; seg < 8; seg++ {
		hi := ip[2*seg]
		lo := ip[2*seg+1]
		out = append(out, lo, hi)
	}
	return out, nil
}
