package hmac_sha256

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func TestExecutor(t *testing.T) {
	e := NewExecutor("sign", []string{
		"111111",
		"appKey",
		"1111111",
		"format",
		"JSON",
		"idcard",
		"111111111111111111",
		"method",
		"realid.idcard.verify",
		"nonce",
		"1111111",
		"realname",
		"张三",
		"signMethod",
		"HMAC-SHA256",
		"signVersion",
		"1",
		"timestamp",
		"2018-02-07 02:50:21",
		"version",
		"1",
	})
	ctx := &Context{}
	v, err := e.Generate(ctx, "")
	t.Log(v, err)
}

type Context struct {
}

func (c *Context) RequestId() string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) AcceptTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Value(key interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func (c *Context) WithValue(key, val interface{}) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Assert(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetLabel(name, value string) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetLabel(name string) string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Labels() map[string]string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetComplete() eocontext.CompleteHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetCompleteHandler(handler eocontext.CompleteHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetFinish() eocontext.FinishHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetFinish(handler eocontext.FinishHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetBalance() eocontext.BalanceHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetBalance(handler eocontext.BalanceHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) RealIP() string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalIP() net.IP {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalPort() int {
	//TODO implement me
	panic("implement me")
}

func (c *Context) IsCloneable() bool {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Clone() (eocontext.EoContext, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Request() http_service.IRequestReader {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Proxy() http_service.IRequest {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Response() http_service.IResponse {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SendTo(scheme string, node eocontext.INode, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Proxies() []http_service.IProxy {
	//TODO implement me
	panic("implement me")
}

func (c *Context) FastFinish() {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetEntry() eosc.IEntry {
	//TODO implement me
	panic("implement me")
}
