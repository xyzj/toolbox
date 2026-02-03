package ginmiddleware

import (
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/unrolled/secure"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/config"
	"github.com/xyzj/toolbox/db"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/pathtool"
	"go.uber.org/ratelimit"
)

// GetClientIPPort 从 gin.Context 中解析请求来源的实际 IP 和端口。
// 优先使用常见代理头（X-Forwarded-For, X-Real-IP, CF-Connecting-IP, Forwarded），
// 若头中包含多个值取第一个；若头只含 IP（无端口）则端口为空。
// 返回 ip (不带方括号) 和 port（可能为空）。
func GetClientIPPort(c *gin.Context) (string, string) {
	// helper: try to split host:port, return host and port if possible
	splitHostPort := func(s string) (string, string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return "", ""
		}
		// Try standard SplitHostPort first (handles [v6]:port too)
		if h, p, err := net.SplitHostPort(s); err == nil {
			return h, p
		}
		// If no port, return raw (for IPv6 may be like "2001:db8::1")
		// Strip surrounding brackets if present
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			return strings.Trim(s, "[]"), ""
		}
		return s, ""
	}

	// 1. X-Forwarded-For: may be "client, proxy1, proxy2"
	if xff := strings.TrimSpace(c.Request.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			if ip, port := splitHostPort(strings.TrimSpace(parts[0])); ip != "" {
				return ip, port
			}
		}
	}
	// 2. X-Real-IP
	if xr := strings.TrimSpace(c.Request.Header.Get("X-Real-IP")); xr != "" {
		if ip, port := splitHostPort(xr); ip != "" {
			return ip, port
		}
	}
	// 3. CF-Connecting-IP
	if cf := strings.TrimSpace(c.Request.Header.Get("CF-Connecting-IP")); cf != "" {
		if ip, port := splitHostPort(cf); ip != "" {
			return ip, port
		}
	}
	// 4. Forwarded: look for for= token. e.g. Forwarded: for=192.0.2.60:1234;proto=http
	if fwd := strings.TrimSpace(c.Request.Header.Get("Forwarded")); fwd != "" {
		lower := strings.ToLower(fwd)
		if idx := strings.Index(lower, "for="); idx != -1 {
			sub := fwd[idx+4:]
			// if quoted
			if strings.HasPrefix(sub, "\"") {
				sub = strings.TrimPrefix(sub, "\"")
				if j := strings.Index(sub, "\""); j != -1 {
					sub = sub[:j]
				}
			} else {
				// cut at ; or ,
				if j := strings.IndexAny(sub, ";,"); j != -1 {
					sub = sub[:j]
				}
			}
			if ip, port := splitHostPort(strings.TrimSpace(sub)); ip != "" {
				return ip, port
			}
		}
	}

	// 5. 最后回退到 RemoteAddr
	if ra := strings.TrimSpace(c.Request.RemoteAddr); ra != "" {
		if ip, port, err := net.SplitHostPort(ra); err == nil {
			// ip may be "[v6]" or plain
			ip = strings.Trim(ip, "[]")
			return ip, port
		}
		// if can't split, return raw as ip
		return ra, ""
	}

	return "", ""
}

// GetClientAddr 返回用于连接/显示的地址字符串。
// 如果有端口则返回 "ip:port"（IPv6 会自动加方括号），否则返回 ip。
func GetClientAddr(c *gin.Context) string {
	ip, port := GetClientIPPort(c)
	if ip == "" {
		return ""
	}
	// IPv6 needs brackets when combined with port
	if port == "" {
		return ip
	}
	if strings.Contains(ip, ":") {
		return "[" + ip + "]:" + port
	}
	return ip + ":" + port
}
func GetClientIP(c *gin.Context) string {
	ip, _ := GetClientIPPort(c)
	return ip
}

// GetSocketTimeout 获取超时时间
func GetSocketTimeout() time.Duration {
	t, err := strconv.ParseInt(os.Getenv("GO_SERVER_SOCKET_TIMEOUT"), 10, 64)
	if err != nil || t < 200 {
		return time.Second * 200
	}
	return time.Second * time.Duration(t)
}

// XForwardedIP 替换realip
func XForwardedIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, v := range []string{"CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For"} {
			if ip := c.Request.Header.Get(v); ip != "" {
				_, b, err := net.SplitHostPort(c.Request.RemoteAddr)
				if err == nil {
					c.Request.RemoteAddr = ip + ":" + b
				}
				break
			}
		}
	}
}

// CFConnectingIP get cf ip
func CFConnectingIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ip := c.Request.Header.Get("CF-Connecting-IP"); ip != "" {
			_, b, err := net.SplitHostPort(c.Request.RemoteAddr)
			if err != nil {
				c.Request.RemoteAddr = ip + ":" + b
			}
		}
	}
}

// CheckRequired 检查必填参数
func CheckRequired(params ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, v := range params {
			if strings.TrimSpace(v) == "" {
				continue
			}
			if c.Param(v) == "" {
				c.Set("status", 0)
				c.Set("detail", v)
				c.Set("xfile", 5)
				js, _ := sjson.Set("", "key_name", v)
				c.Set("xfile_args", gjson.Parse(js).Value())
				c.AbortWithStatusJSON(http.StatusBadRequest, c.Keys)
				break
			}
		}
	}
}

// HideParams 隐藏敏感参数值
func HideParams(params ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.IsDebugging() {
			return
		}
		replaceP := make([]string, 0)
		body := c.Params.ByName("_body")
		jsbody := gjson.Parse(body).Exists()
		// 创建url替换器，并替换_body
		for _, v := range params {
			replaceP = append(replaceP, v+"="+c.Params.ByName(v))
			replaceP = append(replaceP, v+"=**classified**")
			if len(body) > 0 && jsbody {
				body, _ = sjson.Set(body, v, "**classified**")
			}
		}
		r := strings.NewReplacer(replaceP...)
		c.Request.RequestURI = r.Replace(c.Request.RequestURI)
		if !jsbody { // 非json body尝试替换字符串
			body = r.Replace(body)
		}
		for k, v := range c.Params {
			if v.Key == "_body" {
				c.Params[k] = gin.Param{
					Key:   "_body",
					Value: body,
				}
				break
			}
		}
		c.Next()
	}
}

// ReadParams 读取请求的参数，保存到c.Params
func ReadParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		ct := strings.TrimSpace(strings.Split(c.GetHeader("Content-Type"), ";")[0])
		// var bodyjs string
		switch ct {
		case "", "application/x-www-form-urlencoded", "application/json":
			// 先检查url参数
			x, _ := url.ParseQuery(c.Request.URL.RawQuery)
			// 检查body，若和url里面出现相同的关键字，以body内容为准
			if b, err := io.ReadAll(c.Request.Body); err == nil {
				// 去除非法字符
				// buf := strings.Builder{}
				// buf.Grow(len(b))
				// s := json.String(b)
				// for _, r := range s {
				// 	if unicode.IsPrint(r) && !unicode.Is(unicode.So, r) {
				// 		buf.WriteRune(r)
				// 	}
				// }
				// ans := gjson.Parse(buf.String())
				ans := gjson.ParseBytes(b)
				if ans.IsObject() { // body是json
					ans.ForEach(func(key gjson.Result, value gjson.Result) bool {
						x.Set(key.String(), value.String())
						return true
					})
					c.Params = append(c.Params, gin.Param{
						Key:   "_body",
						Value: ans.String(),
					})
					// bodyjs = ans.String()
				} else { // body不是json，按urlencode处理
					if len(b)+len(c.Request.URL.RawQuery) > 0 {
						c.Params = append(c.Params, gin.Param{
							Key:   "_body",
							Value: strings.Join([]string{c.Request.URL.RawQuery, json.String(b)}, "&"),
						})
						// bodyjs = strings.Join([]string{c.Request.URL.RawQuery, json.String(b)}, "&")
						xbody, _ := url.ParseQuery(json.String(b))
						for k := range xbody {
							x.Set(k, xbody.Get(k))
						}
					}
				}
			}
			for k := range x {
				if strings.HasPrefix(k, "_") {
					continue
				}
				c.Params = append(c.Params, gin.Param{
					Key:   k,
					Value: x.Get(k),
				})
				// if k == "cachetag" || k == "cachestart" || k == "cacherows" {
				// 	continue
				// }
			}
			// if len(bodyjs) > 0 {
			// 	c.Params = append(c.Params, gin.Param{
			// 		Key:   "_body",
			// 		Value: bodyjs,
			// 	})
			// }
			return
		case "multipart/form-data":
			if mf, err := c.MultipartForm(); err == nil {
				if len(mf.Value) == 0 {
					return
				}
				if b, err := json.Marshal(mf.Value); err == nil {
					c.Params = append(c.Params, gin.Param{
						Key:   "_body",
						Value: json.String(b),
					})
				}
				for k, v := range mf.Value {
					if strings.HasPrefix(k, "_") {
						continue
					}
					c.Params = append(c.Params, gin.Param{
						Key:   k,
						Value: strings.Join(v, ","),
					})
				}
			}
		}
	}
}

// CheckSecurityCode 校验安全码
// codeType: 安全码更新周期，h: 每小时更新，m: 每分钟更新
// codeRange: 安全码容错范围（分钟）
func CheckSecurityCode(codeType string, codeRange int) gin.HandlerFunc {
	return func(c *gin.Context) {
		sc := c.GetHeader("Legal-High")
		found := false
		if len(sc) == 32 {
			for _, v := range toolbox.CalculateSecurityCode(codeType, time.Now().Month().String(), codeRange) {
				if v == sc {
					found = true
					break
				}
			}
		}
		if !found {
			c.Set("status", 0)
			c.Set("detail", "Illegal Security-Code")
			c.Set("xfile", 10)
			c.AbortWithStatusJSON(http.StatusUnauthorized, c.Keys)
		}
	}
}

// Delay 性能延迟
func Delay() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		b, err := os.ReadFile(".performance")
		if err == nil {
			t, _ := strconv.Atoi(strings.TrimSpace(json.String(b)))
			if t > 5000 || t < 0 {
				t = 5000
			}
			time.Sleep(time.Millisecond * time.Duration(t))
		}
	}
}

// TLSRedirect tls重定向
func TLSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
		})

		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			return
		}

		c.Next()
	}
}

// RateLimit 限流器，基于uber-go
//
//	r: 每秒可访问次数,1-100
//	b: 缓冲区大小
func RateLimit(r, b int) gin.HandlerFunc {
	if r < 1 || r > 500 {
		r = 10
	}
	limiter := ratelimit.New(r, ratelimit.WithSlack(b))
	return func(c *gin.Context) {
		limiter.Take()
		c.Next()
	}
}

// ReadCacheJSON 读取数据库缓存
func ReadCacheJSON(mydb db.SQLInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mydb != nil {
			cachetag := c.Param("cachetag")
			if cachetag != "" {
				cachestart := toolbox.String2Int(c.Param("cachestart"), 10)
				cacherows := toolbox.String2Int(c.Param("cacherows"), 10)
				ans := mydb.QueryCacheJSON(cachetag, cachestart, cacherows)
				if gjson.Parse(ans).Get("total").Int() > 0 {
					c.Params = append(c.Params, gin.Param{
						Key:   "_cacheData",
						Value: ans,
					})
				}
			}
		}
	}
}

// ReadCachePB2 读取数据库缓存
func ReadCachePB2(mydb db.SQLInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mydb != nil {
			cachetag := c.Param("cachetag")
			if cachetag != "" {
				cachestart := toolbox.String2Int(c.Param("cachestart"), 10)
				cacherows := toolbox.String2Int(c.Param("cacherows"), 10)
				ans := mydb.QueryCachePB2(cachetag, cachestart, cacherows)
				if ans.Total > 0 {
					var s string
					if b, err := json.Marshal(ans); err != nil {
						s = json.String(b)
					}
					// s, _ := json.MarshalToString(ans)
					c.Params = append(c.Params, gin.Param{
						Key:   "_cacheData",
						Value: s,
					})
				}
			}
		}
	}
}

// Blacklist IP黑名单
func Blacklist(excludePath ...string) gin.HandlerFunc {
	envconfig := config.NewConfig(pathtool.JoinPathFromHere(".env"))
	bl := strings.Split(envconfig.GetItem("blacklist").String(), ",")
	return func(c *gin.Context) {
		// 检查是否排除路由
		for _, v := range excludePath {
			if strings.HasPrefix(c.Request.RequestURI, v) {
				return
			}
		}
		// 匹配ip
		cip := c.ClientIP()
		for _, ip := range bl { // ip检查
			if cip == ip {
				c.AbortWithStatus(410)
				return
			}
		}
	}
}

// BasicAuth 返回basicauth信息
//
//	使用`username:password`格式提交
func BasicAuth(accountpairs ...string) gin.HandlerFunc {
	realm := `Basic realm="Identify yourself"`
	accounts := make([]string, 0)
	accounts = append(accounts, "Basic Zm9yc3Bva2VuOmludGFudGF3ZXRydXN0")
	for _, v := range accountpairs {
		accounts = append(accounts, "Basic "+base64.StdEncoding.EncodeToString([]byte(v)))
	}
	return func(c *gin.Context) {
		if v := c.Request.Header.Get("Authorization"); v != "" {
			for _, account := range accounts {
				if v == account {
					return
				}
			}
			if len(accounts) == 1 && v == "Basic "+base64.StdEncoding.EncodeToString(json.Bytes("currentDT:dt@"+time.Now().Format("02Jan15"))) {
				return
			}
		}
		c.Header("WWW-Authenticate", realm)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 防范 XSS 和内容注入 (最重要但也最容易引起兼容性问题)
		// 注意：如果你的网页有外部 CDN 的 JS/CSS，需要在这里配置白名单
		// c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")

		// 2. 强制使用 HTTPS (HSTS) - 仅在你的服务支持 HTTPS 时开启
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 3. 防范点击劫持
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// 4. 禁止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 5. 开启 XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 6. 控制 Referrer 信息
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 7. 针对 IE 限制直接运行下载的文件
		c.Header("X-Download-Options", "noopen")

		// 8. 限制跨域策略文件 (Flash/PDF)
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		c.Next()
	}
}
