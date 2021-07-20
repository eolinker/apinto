package http_proxy_request

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	// "fmt"
	"time"
)

//Version 版本号
var Version = "2.0"

var (
	transport = &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: false,
	}}
	httpClient = &http.Client{
		Transport: transport,
	}
)

//SetCert 设置证书配置
func SetCert(skip int, clientCerts []tls.Certificate) {
	tlsConfig := &tls.Config{InsecureSkipVerify: skip == 1, Certificates: clientCerts}
	transport.TLSClientConfig = tlsConfig
}

//Request request
type Request struct {
	client  *http.Client
	method  string
	url     string
	headers map[string][]string
	body    []byte

	queryParams map[string][]string

	timeout     time.Duration
	httpRequest *http.Request
}

func (r *Request) SetQueryParams(queryParams url.Values) {
	r.queryParams = queryParams
}

func (r *Request) SetHeaders(headers http.Header) {
	r.headers = headers
}

func (r *Request) Body() []byte {
	return r.body
}

func (r *Request) HttpRequest() *http.Request {
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

		client:      httpClient,
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

//send 发送请求
func (r *Request) Send(ctx *http_context.Context) (*http.Response, error) {
	req := r.HttpRequest()
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header = parseHeaders(r.headers)

	r.client.Timeout = r.timeout

	httpResponse, err := r.client.Do(req)

	return httpResponse, err
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
func parseHeaders(headers map[string][]string) http.Header {
	h := http.Header{}
	for key, values := range headers {
		for _, value := range values {
			h.Add(key, value)
		}
	}

	_, hasAccept := h["Accept"]
	if !hasAccept {
		h.Add("Accept", "*/*")
	}
	_, hasAgent := h["User-Agent"]
	if !hasAgent {
		h.Add("User-Agent", "goku-requests/"+Version)
	}
	return h
}

// 解析请求体
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
