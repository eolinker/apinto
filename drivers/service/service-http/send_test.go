package service_http

import (
	"testing"

	upstream_http "github.com/eolinker/apinto/drivers/upstream/upstream-http"
	"github.com/eolinker/apinto/upstream/balance"
)

func TestSend(t *testing.T) {
	_ = upstream_http.NewFactory()
	balanceFactory, err := balance.GetFactory("")
	if err != nil {
		t.Error(err)
		return
	}

	anonymous, err := defaultDiscovery.GetApp("www.baidu.com")
	if err != nil {
		t.Error(err)
		return
	}
	balanceHandler, err := balanceFactory.Create(anonymous)
	if err != nil {
		t.Error(err)
		return
	}
	node, _ := balanceHandler.Next()
	t.Log(node.Addr())
}
