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
	DoRequest(*http.Request, ...ReqOptions) (int, []byte, map[string]string, error)
	DoStreamRequest(*http.Request, func(map[string]string), func([]byte) bool, ...ReqOptions) error
}

const (
	HEADER_RESP_FROM     = "Resp-From"
	HEADER_RESP_DURATION = "Resp-Duration"
	HEADER_CONTENT_TYPE  = "Content-Type"
	HEADER_COMPRESSED    = "Compressed"
	HEADER_VALUE_URLE    = "application/x-www-form-urlencoded"
	HEADER_VALUE_JSON    = "application/json; charset=utf-8"
	HEADER_VALUE_ZSTD    = "zstd"
	LogFormater          = "[req] |%d| %-13s |%s %s > %s"
	LogErrFormater       = "[req] |%d| %s %s > %s"
)

type httpOpt struct {
	tls  *tls.Config
	logg logger.Logger
}
type HTTPOptions func(opt *httpOpt)

// WithTLS returns an HTTPOptions function that sets the TLS configuration for the HTTP client.
// The provided tls.Config will be used for secure connections.
func WithTLS(t *tls.Config) HTTPOptions {
	return func(o *httpOpt) {
		o.tls = t
	}
}

// WithLogger returns an HTTPOptions function that sets the logger for HTTP requests.
// It allows customization of logging behavior by injecting a logger.Logger instance
// into the HTTP client options.
func WithLogger(l logger.Logger) HTTPOptions {
	return func(o *httpOpt) {
		o.logg = l
	}
}

var defaultReqOpt = reqOptions{timeout: time.Second * 10}

type reqOptions struct {
	timeout time.Duration
	logreq  bool
}
type ReqOptions func(opt *reqOptions)

// WithLogRequest returns a ReqOptions function that enables logging of the HTTP request.
// When applied, it sets the logreq field of ReqOptions to true.
func WithLogRequest() ReqOptions {
	return func(o *reqOptions) {
		o.logreq = true
	}
}

// WithTimeout returns a ReqOptions function that sets the timeout duration for an HTTP request.
// The provided duration 't' will be applied to the ReqOptions's timeout field.
func WithTimeout(t time.Duration) ReqOptions {
	return func(o *reqOptions) {
		o.timeout = t
	}
}

type Client struct {
	client *http.Client
	logg   logger.Logger
	opt    *reqOptions
}

func (c *Client) ensureRequestOpts(opts ...ReqOptions) {
	opt := defaultReqOpt
	for _, o := range opts {
		o(&opt)
	}
	if c.logg == nil {
		opt.logreq = false
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
func (c *Client) makeRequest(req *http.Request, opts ...ReqOptions) (*http.Request, context.CancelFunc) {
	c.ensureRequestOpts(opts...)
	// Set default content type if not provided
	if req.Header.Get(HEADER_CONTENT_TYPE) == "" {
		switch req.Method {
		case "GET":
			req.Header.Set(HEADER_CONTENT_TYPE, HEADER_VALUE_URLE)
		case "POST":
			req.Header.Set(HEADER_CONTENT_TYPE, HEADER_VALUE_JSON)
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
func (c *Client) DoStreamRequest(req *http.Request, header func(map[string]string), recv func([]byte) error, opts ...ReqOptions) error {
	req, cancel := c.makeRequest(req, opts...)
	defer cancel()
	start := time.Now()
	// Send the request with the timeout context
	resp, err := c.client.Do(req)
	if err != nil {
		c.logg.Error(fmt.Sprintf(LogErrFormater, 500, req.Method, req.URL.String(), err.Error()))
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		c.logg.Error(fmt.Sprintf(LogErrFormater, resp.StatusCode, req.Method, req.URL.String(), json.String(b)))
		return err
	}
	// Return headers
	if header != nil {
		// Collect response headers
		h := make(map[string]string)
		h[HEADER_RESP_FROM] = req.URL.Host
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
func (c *Client) DoRequest(req *http.Request, opts ...ReqOptions) (int, []byte, map[string]string, error) {
	req, cancel := c.makeRequest(req, opts...)
	defer cancel()
	start := time.Now()
	// Send the request with the timeout context
	resp, err := c.client.Do(req)
	if err != nil {
		c.logg.Error(fmt.Sprintf(LogErrFormater, 500, req.Method, req.URL.String(), err.Error()))
		return 500, nil, nil, err
	}
	defer resp.Body.Close()
	sc := resp.StatusCode
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logg.Error(fmt.Sprintf(LogErrFormater, sc, req.Method, req.URL.String(), err.Error()))
		return sc, nil, nil, err
	}

	// Collect response headers
	h := make(map[string]string)
	h[HEADER_RESP_FROM] = req.URL.Host
	h[HEADER_RESP_DURATION] = time.Since(start).String()
	for k := range resp.Header {
		h[k] = resp.Header.Get(k)
	}
	// 日志
	if c.opt.logreq {
		c.logg.Info(fmt.Sprintf(LogFormater, sc, h[HEADER_RESP_DURATION], req.Method, req.URL.String(), json.String(b)))
	}
	return sc, b, h, nil
}

// New creates and returns a new HTTP client with customizable options.
// It accepts a variadic list of HTTPOpts, which are functions that modify the default HTTPOpt configuration.
// The returned Client is initialized with sensible defaults, including a custom TLS configuration,
// connection pooling settings, and a logger. Options provided via HTTPOpts can override these defaults.
func New(opts ...HTTPOptions) *Client {
	opt := &httpOpt{
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
				MaxConnsPerHost:     777,
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				TLSClientConfig:     opt.tls,
			},
		},
		logg: opt.logg,
	}
}
