package http_context

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.IProxy = (*requestAgent)(nil)

type requestAgent struct {
	http_service.IRequest
	host           string
	scheme         string
	statusCode     int
	status         string
	responseLength int
	responseBody   string
	beginTime      time.Time
	endTime        time.Time
	hostAgent      *UrlAgent
	remoteIP       string
	remotePort     int
	originHeader   *fasthttp.ResponseHeader
	headers        http.Header
}

func (a *requestAgent) ResponseBody() string {
	return a.responseBody
}

func (a *requestAgent) ResponseHeaders() http.Header {
	if a.headers != nil {
		return a.headers
	}
	if a.originHeader == nil {
		return make(http.Header)
	}
	headers := make(http.Header)
	a.originHeader.VisitAll(func(key, value []byte) {
		bytes.SplitN(value, []byte(":"), 2)
		headers[string(key)] = []string{string(value)}
	})
	a.headers = headers
	return a.headers
}

func (a *requestAgent) ProxyTime() time.Time {
	return a.beginTime
}

func (a *requestAgent) StatusCode() int {
	return a.statusCode
}

func (a *requestAgent) Status() string {
	return a.status
}

func (a *requestAgent) setStatusCode(code int) {
	a.statusCode = code
	a.status = strconv.Itoa(code)
}

func (a *requestAgent) ResponseLength() int {
	return a.responseLength
}

func (a *requestAgent) setResponseLength(length int) {
	if length > 0 {
		a.responseLength = length
	}
}

func (a *requestAgent) setRemoteIP(ip string) {
	a.remoteIP = ip
}

func (a *requestAgent) setRemotePort(port int) {
	a.remotePort = port
}

func newRequestAgent(IRequest http_service.IRequest, host string, scheme string, header *fasthttp.ResponseHeader, beginTime, endTime time.Time) *requestAgent {
	return &requestAgent{IRequest: IRequest, host: host, scheme: scheme, beginTime: beginTime, endTime: endTime, originHeader: header}
}

func (a *requestAgent) ResponseTime() int64 {
	return a.endTime.Sub(a.beginTime).Milliseconds()
}

func (a *requestAgent) RemoteIP() string {
	return a.remoteIP
}

func (a *requestAgent) RemotePort() int {
	return a.remotePort
}

func (a *requestAgent) URI() http_service.IURIWriter {
	if a.hostAgent == nil {
		a.hostAgent = NewUrlAgent(a.IRequest.URI(), a.host, a.scheme)
	}
	return a.hostAgent
}

type UrlAgent struct {
	http_service.IURIWriter
	host   string
	scheme string
}

func (u *UrlAgent) SetScheme(scheme string) {
	u.scheme = scheme
}
func (u *UrlAgent) Scheme() string {
	return u.scheme
}

func (u *UrlAgent) Host() string {
	return u.host
}

func (u *UrlAgent) SetHost(host string) {
	u.host = host
}

func NewUrlAgent(IURIWriter http_service.IURIWriter, host string, scheme string) *UrlAgent {
	return &UrlAgent{IURIWriter: IURIWriter, host: host, scheme: scheme}
}
