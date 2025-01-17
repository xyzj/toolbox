package toolbox

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	json "github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
)

type RequsetOpts string

var RequestNotLog RequsetOpts = "notlog"

type HttpOpts struct {
	Tls  *tls.Config
	Logg logger.Logger
}

var httpClient = NewHTTPClient(nil)

// DoRequestWithTimeout 发起请求
func DoRequestWithTimeout(req *http.Request, timeo time.Duration, opts ...RequsetOpts) (int, []byte, map[string]string, error) {
	return httpClient.DoRequest(req, timeo, opts...)
}

type Client struct {
	client *http.Client
	logg   logger.Logger
}

// DoRequest sends an HTTP request with the provided parameters and returns the response status code, body, headers, and any error encountered.
//
// Parameters:
// - req: The http.Request object containing the request details.
// - timeo: The timeout duration for the request.
// - opts: Optional request options. Currently, only supports RequestNotLog to disable logging for the request.
//
// Return values:
// - statusCode: The HTTP status code of the response.
// - body: The response body as a byte slice.
// - headers: The response headers as a map of strings.
// - err: Any error encountered during the request or response handling.
func (c *Client) DoRequest(req *http.Request, timeo time.Duration, opts ...RequsetOpts) (int, []byte, map[string]string, error) {
	// 处理头
	if req.Header.Get("Content-Type") == "" {
		switch req.Method {
		case "GET":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case "POST":
			req.Header.Set("Content-Type", "application/json")
		}
	}
	// 超时
	ctx, cancel := context.WithTimeout(context.Background(), timeo)
	defer cancel()
	// 请求
	start := time.Now()
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		c.logg.Error("REQ ERR:" + fmt.Sprintf("%s %s▸%s", req.Method, req.URL.String(), err.Error()))
		return 502, nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logg.Error("REQ ERR:" + fmt.Sprintf("%s %s▸%s", req.Method, req.URL.String(), err.Error()))
		return 502, nil, nil, err
	}
	sc := resp.StatusCode
	end := time.Since(start).String()
	notlog := false
	for _, o := range opts {
		if o == RequestNotLog {
			notlog = true
		}
	}
	// 日志
	if !notlog {
		c.logg.Info("REQ:" + fmt.Sprintf("|%d| %-13s |%s %s ▸%s", sc, end, req.Method, req.URL.String(), json.String(b)))
	}
	// 处理头
	h := make(map[string]string)
	h["Resp-From"] = req.Host
	h["Resp-Duration"] = end
	for k := range resp.Header {
		h[k] = resp.Header.Get(k)
	}
	return sc, b, h, nil
}

func NewHTTPClient(opts *HttpOpts) *Client {
	if opts == nil {
		opts = &HttpOpts{
			Tls:  nil,
			Logg: &logger.NilLogger{},
		}
	}
	if opts.Tls == nil {
		opts.Tls = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				IdleConnTimeout:     time.Second * 10,
				MaxConnsPerHost:     77,
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				TLSClientConfig:     opts.Tls,
			},
		},
		logg: opts.Logg,
	}
}
