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
	n := &nacos{
		client:   newClient(cfg.Config.Address, cfg.getParams()),
		nodes:    discovery.NewNodesData(),
		services: discovery.NewServices(),
		locker:   sync.RWMutex{},
	}
	app, err := n.GetApp(serviceName)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range app.Nodes() {
		t.Log(node.ID())
	}
	ns, bo := n.nodes.Get(serviceName)
	if bo {
		t.Log(len(ns))
	} else {
		t.Error("nodes error")
	}

}
