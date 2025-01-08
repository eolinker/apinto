package script_handler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eolinker/apinto/drivers"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var scriptEaxample = `
package test

import (
"fmt"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
) 

func RenameHost(ctx http_service.IHttpContext) error{
  println("test script")
panic("test")
return fmt.Errorf("test error")
	ctx.Proxy().URI().SetHost("http://testUri.com")
}
`

func TestScript(t *testing.T) {
	var config = &Config{
		Script:  scriptEaxample,
		Package: "test",
		Fname:   "RenameHost",
	}
	fn, err := getFunc(config)
	if err != nil {
		t.Fatal(err)
	}
	var script = Script{
		WorkerBase: drivers.Worker("test", "test"),
		fn:         fn,
	}

	err = script.DoHttpFilter(HttpContext{}, MockChain{})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("test ok")
}

type MockChain struct{}

func (mc MockChain) DoChain(ctx eocontext.EoContext) error {
	return nil
}
func (mc MockChain) Destroy() {

}

type HttpContext struct {
}

// AcceptTime implements http_context.IHttpContext.
func (h HttpContext) AcceptTime() time.Time {
	panic("unimplemented")
}

// Assert implements http_context.IHttpContext.
func (h HttpContext) Assert(i interface{}) error {
	panic("unimplemented")
}

// Clone implements http_context.IHttpContext.
func (h HttpContext) Clone() (eocontext.EoContext, error) {
	panic("unimplemented")
}

// Context implements http_context.IHttpContext.
func (h HttpContext) Context() context.Context {
	panic("unimplemented")
}

// FastFinish implements http_context.IHttpContext.
func (h HttpContext) FastFinish() {
	panic("unimplemented")
}

// GetBalance implements http_context.IHttpContext.
func (h HttpContext) GetBalance() eocontext.BalanceHandler {
	panic("unimplemented")
}

// GetComplete implements http_context.IHttpContext.
func (h HttpContext) GetComplete() eocontext.CompleteHandler {
	panic("unimplemented")
}

// GetEntry implements http_context.IHttpContext.
func (h HttpContext) GetEntry() eosc.IEntry {
	panic("unimplemented")
}

// GetFinish implements http_context.IHttpContext.
func (h HttpContext) GetFinish() eocontext.FinishHandler {
	panic("unimplemented")
}

// GetLabel implements http_context.IHttpContext.
func (h HttpContext) GetLabel(name string) string {
	panic("unimplemented")
}

// GetUpstreamHostHandler implements http_context.IHttpContext.
func (h HttpContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	panic("unimplemented")
}

// IsCloneable implements http_context.IHttpContext.
func (h HttpContext) IsCloneable() bool {
	panic("unimplemented")
}

// Labels implements http_context.IHttpContext.
func (h HttpContext) Labels() map[string]string {
	panic("unimplemented")
}

// LocalAddr implements http_context.IHttpContext.
func (h HttpContext) LocalAddr() net.Addr {
	panic("unimplemented")
}

// LocalIP implements http_context.IHttpContext.
func (h HttpContext) LocalIP() net.IP {
	panic("unimplemented")
}

// LocalPort implements http_context.IHttpContext.
func (h HttpContext) LocalPort() int {
	panic("unimplemented")
}

// Proxies implements http_context.IHttpContext.
func (h HttpContext) Proxies() []http_context.IProxy {
	panic("unimplemented")
}

// Proxy implements http_context.IHttpContext.
func (h HttpContext) Proxy() http_context.IRequest {
	return Request{}
}

// RealIP implements http_context.IHttpContext.
func (h HttpContext) RealIP() string {
	panic("unimplemented")
}

// Request implements http_context.IHttpContext.
func (h HttpContext) Request() http_context.IRequestReader {
	panic("unimplemented")
}

// RequestId implements http_context.IHttpContext.
func (h HttpContext) RequestId() string {
	panic("unimplemented")
}

// Response implements http_context.IHttpContext.
func (h HttpContext) Response() http_context.IResponse {
	panic("unimplemented")
}

// Scheme implements http_context.IHttpContext.
func (h HttpContext) Scheme() string {
	panic("unimplemented")
}

// SendTo implements http_context.IHttpContext.
func (h HttpContext) SendTo(scheme string, node eocontext.INode, timeout time.Duration) error {
	panic("unimplemented")
}

// SetBalance implements http_context.IHttpContext.
func (h HttpContext) SetBalance(handler eocontext.BalanceHandler) {
	panic("unimplemented")
}

// SetCompleteHandler implements http_context.IHttpContext.
func (h HttpContext) SetCompleteHandler(handler eocontext.CompleteHandler) {
	panic("unimplemented")
}

// SetFinish implements http_context.IHttpContext.
func (h HttpContext) SetFinish(handler eocontext.FinishHandler) {
	panic("unimplemented")
}

// SetLabel implements http_context.IHttpContext.
func (h HttpContext) SetLabel(name string, value string) {
	panic("unimplemented")
}

// SetUpstreamHostHandler implements http_context.IHttpContext.
func (h HttpContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	panic("unimplemented")
}

// Value implements http_context.IHttpContext.
func (h HttpContext) Value(key interface{}) interface{} {
	panic("unimplemented")
}

// WithValue implements http_context.IHttpContext.
func (h HttpContext) WithValue(key interface{}, val interface{}) {
	panic("unimplemented")
}

type Request struct {
}

// Body implements http_context.IRequest.
func (r Request) Body() http_context.IBodyDataWriter {
	panic("unimplemented")
}

// ContentLength implements http_context.IRequest.
func (r Request) ContentLength() int {
	panic("unimplemented")
}

// ContentType implements http_context.IRequest.
func (r Request) ContentType() string {
	return "application/json; charset=utf-8"
}

// Header implements http_context.IRequest.
func (r Request) Header() http_context.IHeaderWriter {
	panic("unimplemented")
}

// Method implements http_context.IRequest.
func (r Request) Method() string {
	panic("unimplemented")
}

// SetMethod implements http_context.IRequest.
func (r Request) SetMethod(method string) {
	panic("unimplemented")
}

// URI implements http_context.IRequest.
func (r Request) URI() http_context.IURIWriter {
	return URIWriter{}
}

type URIWriter struct {
	host string
}

func (writer URIWriter) SetPath(string)          {}
func (writer URIWriter) SetScheme(scheme string) {}
func (writer URIWriter) SetHost(host string) {
	writer.host = host
}
func (writer URIWriter) RequestURI() string         { return "" }
func (writer URIWriter) Scheme() string             { return "" }
func (writer URIWriter) RawURL() string             { return "" }
func (writer URIWriter) Host() string               { return writer.host }
func (writer URIWriter) Path() string               { return "" }
func (writer URIWriter) SetQuery(key, value string) {}
func (writer URIWriter) AddQuery(key, value string) {}
func (writer URIWriter) DelQuery(key string)        {}
func (writer URIWriter) SetRawQuery(raw string)     {}
func (writer URIWriter) GetQuery(key string) string { return "" }
func (writer URIWriter) RawQuery() string           { return "" }
