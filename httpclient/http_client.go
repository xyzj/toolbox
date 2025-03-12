package httpclient

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	json "github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
)

type HTTPClient interface {
	DoRequest(*http.Request, ...ReqOpts) (int, []byte, map[string]string, error)
	DoStreamRequest(*http.Request, func(map[string]string), func([]byte) bool, ...ReqOpts) error
}

const (
	HEADER_RESP_FROM     = "Resp-From"
	HEADER_RESP_DURATION = "Resp-Duration"
)

type HTTPOpt struct {
	tls  *tls.Config
	logg logger.Logger
}
type HTTPOpts func(opt *HTTPOpt)

func OptTLS(t *tls.Config) HTTPOpts {
	return func(o *HTTPOpt) {
		o.tls = t
	}
}

func OptLogger(l logger.Logger) HTTPOpts {
	return func(o *HTTPOpt) {
		o.logg = l
	}
}

var defaultReqOpt = ReqOpt{timeout: time.Second * 10}

type ReqOpt struct {
	timeout time.Duration
	notLog  bool
}
type ReqOpts func(opt *ReqOpt)

func OptNotLog() ReqOpts {
	return func(o *ReqOpt) {
		o.notLog = true
	}
}

func OptTimeout(t time.Duration) ReqOpts {
	return func(o *ReqOpt) {
		o.timeout = t
	}
}

type Client struct {
	client *http.Client
	logg   logger.Logger
	opt    *ReqOpt
}

func (c *Client) ensureRequestOpts(opts ...ReqOpts) {
	opt := defaultReqOpt
	for _, o := range opts {
		o(&opt)
	}
	if c.logg == nil {
		opt.notLog = true
	}
	c.opt = &opt
}

// makeRequest prepares an HTTP request with the provided options and returns the modified request along with a context cancellation function.
// It sets a default content type based on the request method if not provided, and applies the specified request options.
//
// Parameters:
// - req: The http.Request object containing the request details.
// - opts: Variadic ReqOpts parameters that specify optional request settings.
//
// Return values:
// - *http.Request: The modified http.Request object with applied options and a timeout context.
// - context.CancelFunc: A function to cancel the timeout context.
func (c *Client) makeRequest(req *http.Request, opts ...ReqOpts) (*http.Request, context.CancelFunc) {
	c.ensureRequestOpts(opts...)
	// Set default content type if not provided
	if req.Header.Get("Content-Type") == "" {
		switch req.Method {
		case "GET":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case "POST":
			req.Header.Set("Content-Type", "application/json")
		}
	}
	timeoCtx, cancel := context.WithTimeout(context.Background(), c.opt.timeout)
	return req.WithContext(timeoCtx), cancel
}

// DoStreamRequest sends an HTTP request with streaming capabilities and processes the response in chunks.
// It handles timeouts, logs the request and response details, and allows for custom header and data processing.
//
// Parameters:
// - req: The http.Request object containing the request details.
// - header: A function that receives the response headers as a map of strings.
// - recv: A function that processes each chunk of the response body. Return false to stop processing.
// - opts: Optional request options. Currently, only supports OptNotLog to disable logging for the request.
//
// Return value:
// - An error if any occurred during the request or response handling.
func (c *Client) DoStreamRequest(req *http.Request, header func(map[string]string), recv func([]byte) error, opts ...ReqOpts) error {
	req, cancel := c.makeRequest(req, opts...)
	defer cancel()
	start := time.Now()
	// Send the request with the timeout context
	resp, err := c.client.Do(req)
	if err != nil {
		c.logg.Error("Request error:" + fmt.Sprintf("%s %s>%s", req.Method, req.URL.String(), err.Error()))
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		c.logg.Error("Response status code not OK:" + fmt.Sprintf("%s %s>%d,%s", req.Method, req.URL.String(), resp.StatusCode, string(b)))
		return err
	}
	// Return headers
	if header != nil {
		// Collect response headers
		h := make(map[string]string)
		h[HEADER_RESP_FROM] = req.Host
		h[HEADER_RESP_DURATION] = time.Since(start).String()
		for k := range resp.Header {
			h[k] = resp.Header.Get(k)
		}
		header(h)
	}
	if recv != nil {
		// buf := bufio.NewReader(resp.Body)
		// b := make([]byte, 65536)
		var bb []byte
		// var n int
		// var err error
		// for {
		// 	if c.opt.readBytes == 0 {
		// 		n, err = buf.Read(b)
		// 		if n > 0 {
		// 			bb = b[:n]
		// 		}
		// 	} else {
		// 		bb, err = buf.ReadBytes(c.opt.readBytes)
		// 		bb = append(bb, c.opt.readBytes)
		// 	}
		// 	if err == io.EOF {
		// 		break
		// 	}
		// 	if err != nil {
		// 		c.logg.Error("Read response body error:" + err.Error())
		// 		return err
		// 	}
		// 	if err = recv(bb); err != nil {
		// 		return err
		// 	}
		// }
		buf := bufio.NewScanner(resp.Body)
		for buf.Scan() {
			bb = buf.Bytes()
			bb = append(bb, '\n')
			if err = recv(bb); err != nil {
				return err
			}
		}
	}
	return nil
}

// DoRequest sends an HTTP request with the provided parameters and returns the response, headers, and any error encountered.
// It handles timeouts, logs the request and response details, and collects response headers.
//
// Parameters:
// - req: The http.Request object containing the request details.
// - opts: Optional request options. Currently, only supports OptNotLog to disable logging for the request.
//
// Return values:
// - statusCode: The HTTP status code of the response.
// - body: The response body as a byte slice.
// - headers: The response headers as a map of strings.
// - err: Any error encountered during the request or response handling.
func (c *Client) DoRequest(req *http.Request, opts ...ReqOpts) (int, []byte, map[string]string, error) {
	req, cancel := c.makeRequest(req, opts...)
	defer cancel()
	start := time.Now()
	// Send the request with the timeout context
	resp, err := c.client.Do(req)
	if err != nil {
		c.logg.Error("REQ ERR:" + fmt.Sprintf("%s %s>%s", req.Method, req.URL.String(), err.Error()))
		return 502, nil, nil, err
	}
	if resp.StatusCode != 200 {
		c.logg.Error("REQ NOT OK:" + fmt.Sprintf("%s %s>%d", req.Method, req.URL.String(), resp.StatusCode))
		return 502, nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logg.Error("RESP READ ERR:" + fmt.Sprintf("%s %s>%s", req.Method, req.URL.String(), err.Error()))
		return 502, nil, nil, err
	}
	sc := resp.StatusCode

	// Collect response headers
	h := make(map[string]string)
	h[HEADER_RESP_FROM] = req.Host
	h[HEADER_RESP_DURATION] = time.Since(start).String()
	for k := range resp.Header {
		h[k] = resp.Header.Get(k)
	}
	// 日志
	if !c.opt.notLog {
		c.logg.Info("REQ:" + fmt.Sprintf("|%d| %-13s |%s %s>%s", sc, h[HEADER_RESP_DURATION], req.Method, req.URL.String(), json.String(b)))
	}
	return sc, b, h, nil
}

func New(opts ...HTTPOpts) *Client {
	opt := &HTTPOpt{
		tls: &tls.Config{
			InsecureSkipVerify: true,
		},
		logg: &logger.NilLogger{},
	}
	for _, o := range opts {
		o(opt)
	}
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				IdleConnTimeout:     time.Second * 10,
				MaxConnsPerHost:     77,
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				TLSClientConfig:     opt.tls,
			},
		},
		logg: opt.logg,
	}
}
