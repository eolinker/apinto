package http_context

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

//RequestReader 请求reader
type RequestReader struct {
	*Header
	*BodyRequestHandler
	req  *http.Request
	body []byte
}

//Proto 获取协议
func (r *RequestReader) Proto() string {
	return r.req.Proto
}

//NewRequestReader 创建RequestReader
func NewRequestReader(req fasthttp.Request) *RequestReader {
	r := new(RequestReader)
	r.ParseRequest(req)
	return r
}

//ParseRequest 解析请求
func (r *RequestReader) ParseRequest(req fasthttp.Request) {

	//newReq, _ := http.NewRequest(string(req.Header.Method()), string(req.URI().FullURI()), nil)
	uri, _ := url.Parse(string(req.URI().FullURI()))
	newReq := &http.Request{
		Method:        string(req.Header.Method()),
		URL:           uri,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		ContentLength: 0,
		Host:          uri.Host,
		RequestURI:    uri.RequestURI(),
	}
	hs := strings.Split(string(req.Header.Header()), "\r\n")
	for i, h := range hs {
		if i == 0 {
			continue
		}
		values := strings.Split(h, ":")
		vLen := len(values)
		if vLen < 2 {
			if values[0] != "" {
				newReq.Header.Set(values[0], "")
			}
		} else {
			newReq.Header.Set(values[0], values[1])
		}
	}
	newReq.URL.RawQuery = string(req.URI().QueryString())
	r.req = newReq
	r.Header = NewHeader(r.req.Header)

	r.BodyRequestHandler = NewBodyRequestHandler(r.req.Header.Get("Content-Type"), req.Body())
}

//Cookie 获取cookie
func (r *RequestReader) Cookie(name string) (*http.Cookie, error) {
	return r.req.Cookie(name)
}

//Cookies 获取cookies
func (r *RequestReader) Cookies() []*http.Cookie {
	return r.req.Cookies()
}

//method 获取请求方式
func (r *RequestReader) Method() string {
	return r.req.Method
}

//url url
func (r *RequestReader) URL() *url.URL {
	return r.req.URL
}

//RequestURI 获取请求URI
func (r *RequestReader) RequestURI() string {
	return r.req.RequestURI
}

//Host 获取host
func (r *RequestReader) Host() string {
	return r.req.Host
}

//RemoteAddr 远程地址
func (r *RequestReader) RemoteAddr() string {
	return r.req.RemoteAddr
}

func (r *RequestReader) Request() *http.Request {
	return r.req
}
