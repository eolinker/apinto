package nacos

import (
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestGetApp(t *testing.T) {
	serviceName := "nacos.naming.serviceName"
	cfg := Config{
		Name:   "nacos",
		Scheme: "http",
		Config: AccessConfig{
			Address: []string{
				"10.1.94.48:8848",
			},
			Params: map[string]string{
				"username": "test",
				"password": "test",
			},
		},
	}
	n := &nacos{
		id:       "1",
		name:     cfg.Name,
		client:   newClient(cfg.Config.Address, cfg.getParams(), cfg.getScheme()),
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
