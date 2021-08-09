package http_proxy_request

import (
	"crypto/tls"

	"github.com/valyala/fasthttp"

	// "fmt"
	"time"
)

//Version 版本号
var Version = "2.0"

var (
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
	timeout     time.Duration
	httpRequest *fasthttp.Request
}

//NewRequest 创建新请求
func NewRequest(req *fasthttp.Request, timeout time.Duration) *Request {
	return &Request{
		timeout:     timeout,
		httpRequest: req,
	}
}

//Send 发送请求
func (r *Request) Send() (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	err := httpClient.DoTimeout(r.httpRequest, resp, r.timeout)

	return resp, err
}
