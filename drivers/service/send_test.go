package service

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	"github.com/eolinker/eosc/eocontext"
)

func TestSend(t *testing.T) {

	balanceFactory, err := balance.GetFactory("")
	if err != nil {
		t.Error(err)
		return
	}

	anonymous, err := defaultHttpDiscovery.GetApp("www.baidu.com")
	if err != nil {
		t.Error(err)
		return
	}
	balanceHandler, err := balanceFactory.Create("", 0)
	if err != nil {
		t.Error(err)
		return
	}

	node, _, _ := balanceHandler.Select(&testAppContext{app: &testApp{app: anonymous}})
	t.Log(node.Addr())
}

type testApp struct {
	app discovery.IApp
}

func (t *testApp) Nodes() []eocontext.INode {
	return t.app.Nodes()
}

func (t *testApp) Scheme() string {
	return "http"
}

func (t *testApp) TimeOut() time.Duration {
	return time.Second
}

type testAppContext struct {
	app eocontext.EoApp
}

func (t *testAppContext) RequestId() string {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) AcceptTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Value(key interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) WithValue(key, val interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Assert(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) SetLabel(name, value string) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) GetLabel(name string) string {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Labels() map[string]string {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) GetComplete() eocontext.CompleteHandler {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) SetCompleteHandler(handler eocontext.CompleteHandler) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) GetFinish() eocontext.FinishHandler {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) SetFinish(handler eocontext.FinishHandler) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) GetApp() eocontext.EoApp {
	return t.app
}

func (t *testAppContext) SetApp(app eocontext.EoApp) {
	t.app = app
}

func (t *testAppContext) GetBalance() eocontext.BalanceHandler {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) SetBalance(handler eocontext.BalanceHandler) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) ReadIP() string {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) LocalIP() net.IP {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) LocalPort() int {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) IsCloneable() bool {
	//TODO implement me
	panic("implement me")
}

func (t *testAppContext) Clone() (eocontext.EoContext, error) {
	//TODO implement me
	panic("implement me")
}
