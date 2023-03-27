package service

import (
	"github.com/eolinker/apinto/upstream/balance"
	"testing"
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
	balanceHandler, err := balanceFactory.Create()
	if err != nil {
		t.Error(err)
		return
	}
	node, _, _ := balanceHandler.Select(anonymous)
	t.Log(node.Addr())
}
