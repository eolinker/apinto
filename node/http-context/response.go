package http_context

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/bytebufferpool"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	ResponseHeader
	*fasthttp.Response
	length          int
	responseTime    time.Duration
	statusCode      int
	status          string
	proxyStatusCode int
	proxyStatus     string
	responseError   error
	remoteIP        string
	remotePort      int
	streamBody      *bytes.Buffer
	streamFuncArray []http_service.StreamFunc
}

func (r *Response) StreamFunc() []http_service.StreamFunc {
	return r.streamFuncArray
}

func (r *Response) AppendStreamFunc(streamFunc http_service.StreamFunc) {
	if r.streamFuncArray == nil {
		r.streamFuncArray = make([]http_service.StreamFunc, 0, 10)
	}
	r.streamFuncArray = append(r.streamFuncArray, streamFunc)
}

func (r *Response) StreamFuncHandle(ctx http_service.IHttpContext, org []byte) ([]byte, error) {
	result := make([]byte, len(org))
	copy(result, org)
	var err error
	for _, streamFunc := range r.streamFuncArray {
		result, err = streamFunc(ctx, result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
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
	r.streamBody.Reset()
	return nil
}
func (r *Response) reset(resp *fasthttp.Response) {
	r.Response = resp
	r.ResponseHeader.reset(&resp.Header)
	r.responseError = nil
	r.proxyStatusCode = 0
	if r.streamBody == nil {
		r.streamBody = &bytes.Buffer{}
	}
	r.streamBody.Reset()
}

func (r *Response) BodyLen() int {
	return r.header.ContentLength()
}

func (r *Response) GetBody() []byte {
	if r.IsBodyStream() {
		return r.streamBody.Bytes()
	}
	body, _ := r.BodyUncompressed()
	r.SetHeader("Content-Length", strconv.Itoa(len(body)))
	r.DelHeader("Content-Encoding")
	r.Response.SetBody(body)
	return r.Response.Body()
}

func (r *Response) IsBodyStream() bool {
	return r.Response.IsBodyStream() && r.Response.Header.ContentLength() < 0
}

func (r *Response) SetBody(bytes []byte) {
	if r.IsBodyStream() {
		r.streamBody.Write(bytes)
		// 不处理
		return
	}

	if strings.Contains(r.GetHeader("Content-Encoding"), "gzip") {
		var bb bytebufferpool.ByteBuffer
		_, err := fasthttp.WriteGunzip(&bb, bytes)
		if err == nil {
			r.DelHeader("Content-Encoding")
			r.SetHeader("Content-Length", strconv.Itoa(len(bb.B)))
			r.Response.SetBody(bb.B)
			r.responseError = nil
			return
		}
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
	return strconv.Itoa(r.Response.StatusCode())
}

func (r *Response) SetStatus(code int, status string) {
	r.Response.SetStatusCode(code)
	r.responseError = nil
}

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
