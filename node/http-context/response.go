package http_context

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	ResponseHeader
	*fasthttp.Response
	length          int
	responseTime    time.Duration
	proxyStatusCode int
	responseError   error
	remoteIP        string
	remotePort      int
	bodyStream      *BodyStream
}

type BodyStream struct {
	reader             io.Reader
	streamReadHandler  []func(p []byte) error
	streamWriteHandler []func(p []byte) ([]byte, error)
}

func NewBodyStream(reader io.Reader) *BodyStream {
	return &BodyStream{reader: reader}
}

func (b *BodyStream) AppendReaderFunc(f func(p []byte) error) {
	if b.streamReadHandler == nil {
		b.streamReadHandler = make([]func(p []byte) error, 0)
	}
	b.streamReadHandler = append(b.streamReadHandler, f)
}

func (b *BodyStream) AppendWriterFunc(f func(p []byte) ([]byte, error)) {
	if b.streamWriteHandler == nil {
		b.streamWriteHandler = make([]func(p []byte) ([]byte, error), 0)
	}
	b.streamWriteHandler = append(b.streamWriteHandler, f)
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 默认 32KB 缓冲区
	},
}

func (b *BodyStream) Read(p []byte) (n int, err error) {
	tmp := bufferPool.Get().([]byte)
	defer bufferPool.Put(tmp)
	n, err = b.reader.Read(tmp)
	if err != nil {
		return 0, err
	}
	org := tmp[:n]
	for _, fn := range b.streamWriteHandler {
		result, err := fn(org)
		if err != nil {
			return 0, err
		}
		org = result
	}
	org = append(org, []byte("\n")...)
	copy(p, org)
	return len(org), nil
}

func (b *BodyStream) Write(p []byte) (n int, err error) {

	return 0, nil
}

func (r *Response) GetBodyStream() http_service.IResponseStream {
	return r.bodyStream
}

func (r *Response) ContentLength() int {
	if r.length == 0 {
		return r.Response.Header.ContentLength()
	}
	return r.length
}

func (r *Response) ContentType() string {
	return string(r.Response.Header.ContentType())
}

func (r *Response) HeadersString() string {
	return r.header.String()
}

func (r *Response) ResponseError() error {
	return r.responseError
}

func (r *Response) ClearError() {
	r.responseError = nil
}
func (r *Response) Finish() error {
	r.ResponseHeader.Finish()
	r.Response = nil
	r.responseError = nil
	r.proxyStatusCode = 0
	return nil
}
func (r *Response) reset(resp *fasthttp.Response) {
	r.Response = resp
	r.ResponseHeader.reset(&resp.Header)
	r.responseError = nil
	r.proxyStatusCode = 0
}

func (r *Response) BodyLen() int {
	return r.header.ContentLength()
}

func (r *Response) GetBody() []byte {
	if strings.Contains(r.GetHeader("Content-Encoding"), "gzip") {
		body, _ := r.BodyGunzip()
		r.DelHeader("Content-Encoding")
		r.SetHeader("Content-Length", strconv.Itoa(len(body)))
		r.Response.SetBody(body)
	}
	if r.IsBodyStream() {
		return nil
	}
	return r.Response.Body()
}

func (r *Response) IsBodyStream() bool {
	return r.Response.IsBodyStream() && r.Response.Header.ContentLength() < 0
}

func (r *Response) SetBody(bytes []byte) {
	if r.IsBodyStream() {
		return
	}
	if strings.Contains(r.GetHeader("Content-Encoding"), "gzip") {
		r.DelHeader("Content-Encoding")
	}
	r.Response.SetBody(bytes)
	r.length = len(bytes)
	r.SetHeader("Content-Length", strconv.Itoa(r.length))
	r.responseError = nil
}

func (r *Response) StatusCode() int {
	if r.responseError != nil {
		return 504
	}
	return r.Response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.StatusCode())
}

func (r *Response) SetStatus(code int, status string) {
	r.Response.SetStatusCode(code)
	r.responseError = nil
}

// 原始的响应状态码
func (r *Response) ProxyStatusCode() int {
	return r.proxyStatusCode
}

func (r *Response) ProxyStatus() string {
	return strconv.Itoa(r.proxyStatusCode)
}

func (r *Response) SetProxyStatus(code int, status string) {
	r.proxyStatusCode = code
}

func (r *Response) SetResponseTime(t time.Duration) {
	r.responseTime = t
}

func (r *Response) ResponseTime() time.Duration {
	return r.responseTime
}

func (r *Response) RemoteIP() string {
	return r.remoteIP
}

func (r *Response) RemotePort() int {
	return r.remotePort
}
