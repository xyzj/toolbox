package toolbox

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

var httpClient = NewHTTPClient()

// DoRequestWithTimeout 发起请求
func DoRequestWithTimeout(req *http.Request, timeo time.Duration) (int, []byte, map[string]string, error) {
	return httpClient.DoRequest(req, timeo)
}

type Client struct {
	client *http.Client
}

// DoRequest 发起请求
func (c *Client) DoRequest(req *http.Request, timeo time.Duration) (int, []byte, map[string]string, error) {
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
		return 502, nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 502, nil, nil, err
	}
	// 处理头
	h := make(map[string]string)
	h["Resp-From"] = req.Host
	h["Resp-Duration"] = time.Since(start).String()
	for k := range resp.Header {
		h[k] = resp.Header.Get(k)
	}
	sc := resp.StatusCode
	return sc, b, h, nil
}

func NewHTTPClient() *Client {
	return NewHTTPClientWithTLS(nil)
}

func NewHTTPClientWithTLS(tlsopt *tls.Config) *Client {
	if tlsopt == nil {
		tlsopt = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return &Client{client: &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			IdleConnTimeout:     time.Second * 10,
			MaxConnsPerHost:     77,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			TLSClientConfig:     tlsopt,
		},
	}}
}
