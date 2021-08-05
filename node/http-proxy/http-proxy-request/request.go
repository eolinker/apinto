package http_proxy_request

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	// "fmt"
	"time"
)

//Version 版本号
var Version = "2.0"

var (
	//transport = &http.Transport{TLSClientConfig: &tls.Config{
	//	InsecureSkipVerify: false,
	//},
	//	DialContext: (&net.Dialer{
	//		Timeout:   30 * time.Second, // 连接超时时间
	//		KeepAlive: 60 * time.Second, // 保持长连接的时间
	//	}).DialContext, // 设置连接的参数
	//	MaxIdleConns:          500,              // 最大空闲连接
	//	IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
	//	ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
	//	MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
	//}

	httpClient = &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		MaxConnsPerHost: 4000,
	}
)

//SetCert 设置证书配置
func SetCert(skip int, clientCerts []tls.Certificate) {
	tlsConfig := &tls.Config{InsecureSkipVerify: skip == 1, Certificates: clientCerts}
	httpClient.TLSConfig = tlsConfig
}

//Request http-proxy 请求结构体
type Request struct {
	method  string
	url     string
	headers map[string][]string
	body    []byte

	queryParams map[string][]string

	timeout     time.Duration
	httpRequest *http.Request
}

//SetQueryParams 替换Query参数
func (r *Request) SetQueryParams(queryParams url.Values) {
	r.queryParams = queryParams
}

//SetHeaders 替换header参数
func (r *Request) SetHeaders(headers http.Header) {
	r.headers = headers
}

//Body 返回请求的body参数
func (r *Request) Body() []byte {
	return r.body
}

//HTTPRequest 返回http请求结构体
func (r *Request) HTTPRequest() *http.Request {
	return r.httpRequest
}

//NewRequest 创建新请求
func NewRequest(method string, URL *url.URL) (*Request, error) {
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" &&
		method != "HEAD" && method != "OPTIONS" && method != "PATCH" {
		return nil, errors.New("Unsupported Request method")
	}
	return newRequest(method, URL)
}

//URLPath urlPath
func URLPath(url string, query url.Values) string {
	if len(query) < 1 {
		return url
	}
	return url + "?" + query.Encode()
}

func newRequest(method string, URL *url.URL) (*Request, error) {
	var urlPath string
	queryParams := make(map[string][]string)
	for key, values := range URL.Query() {
		queryParams[key] = values
	}
	urlPath = URL.Scheme + "://" + URL.Host + URL.Path

	r := &Request{
		method:      method,
		url:         urlPath,
		headers:     make(map[string][]string),
		queryParams: queryParams,
	}
	return r, nil
}

//SetHeader 设置请求头
func (r *Request) SetHeader(key string, values ...string) {
	if len(values) > 0 {
		r.headers[key] = values[:]
	} else {
		delete(r.headers, key)
	}
}

//Headers 获取请求头
func (r *Request) Headers() map[string][]string {
	headers := make(map[string][]string)
	for key, values := range r.headers {
		headers[key] = values[:]
	}
	return headers
}

//SetQueryParam 设置Query参数
func (r *Request) SetQueryParam(key string, values ...string) {
	if len(values) > 0 {
		r.queryParams[key] = values[:]
	} else {
		delete(r.queryParams, key)
	}
}

//SetTimeout 设置请求超时时间
func (r *Request) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

//Send 发送请求
func (r *Request) Send(ctx *http_context.Context) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.Header = parseHeaders(r.headers)
	req.SetRequestURI(r.url)
	req.Header.SetMethod(r.method)
	req.Header.Set("Accept-Encoding", "gzip")

	resp := fasthttp.AcquireResponse()

	err := httpClient.DoTimeout(req, resp, r.timeout)

	return resp, err
}

//QueryParams 获取query参数
func (r *Request) QueryParams() map[string][]string {
	params := make(map[string][]string)
	for key, values := range r.queryParams {
		params[key] = values[:]
	}
	return params
}

//URLPath 获取完整的URL路径
func (r *Request) URLPath() string {
	if len(r.queryParams) > 0 {
		return r.url + "?" + parseParams(r.queryParams).Encode()
	}
	return r.url
}

//SetURL 设置URL
func (r *Request) SetURL(url string) {
	r.url = url
}

//SetRawBody 设置源数据
func (r *Request) SetRawBody(body []byte) {
	r.body = body
}

// 解析请求头
func parseHeaders(headers map[string][]string) fasthttp.RequestHeader {
	h := fasthttp.RequestHeader{}
	hasAccept := false
	hasAgent := false
	for key, values := range headers {
		key = strings.ToLower(key)
		for _, value := range values {
			if key == "accept" {
				hasAccept = true
			}
			if key == "user-agent" {
				hasAgent = true
			}
			h.Add(key, value)
		}
	}

	if !hasAccept {
		h.Add("Accept", "*/*")
	}
	if !hasAgent {
		h.Add("User-Agent", "goku-requests/"+Version)
	}
	return h
}

//ParseBody 解析请求体
func (r *Request) ParseBody() error {
	if r.httpRequest == nil {
		var body io.Reader = nil
		if len(r.body) > 0 {
			body = bytes.NewBuffer(r.body)
		}
		request, err := http.NewRequest(r.method, r.URLPath(), body)
		if err != nil {
			return err
		}
		r.httpRequest = request
	}
	return nil
}

// 解析参数
func parseParams(params map[string][]string) url.Values {
	v := url.Values{}
	for key, values := range params {
		for _, value := range values {
			v.Add(key, value)
		}
	}
	return v
}

// 解析URL
func parseURL(urlPath string) (URL *url.URL, err error) {
	URL, err = url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	if URL.Scheme != "http" && URL.Scheme != "https" {
		urlPath = "http://" + urlPath
		URL, err = url.Parse(urlPath)
		if err != nil {
			return nil, err
		}

		if URL.Scheme != "http" && URL.Scheme != "https" {
			return nil, errors.New("[package requests] only HTTP and HTTPS are accepted")
		}
	}
	return
}
