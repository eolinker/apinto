package nacos

import (
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestGetApp(t *testing.T) {
	serviceName := "demo"
	cfg := Config{
		Config: AccessConfig{
			Address: []string{
				"172.18.166.219:8848",
			},
			Params: map[string]string{
				"namespaceId": "82eab342-52a5-400d-a601-3dd7b7d4029c",
			},
		},
	}
	c, _ := newClient("asd", cfg.Config.Address, cfg.Config.Params)
	n := &nacos{
		client:   c,
		services: discovery.NewAppContainer(),
		locker:   sync.RWMutex{},
	}
	app, err := n.GetApp(serviceName)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range app.Nodes() {
		t.Log(node.ID())
	}
	ns, err := n.GetApp(serviceName)
	if err == nil {
		t.Log(len(ns.Nodes()))
	} else {
		t.Error("nodes error")
	}

}
