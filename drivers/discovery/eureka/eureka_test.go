package eureka

import (
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestGetApp(t *testing.T) {
	serviceName := "DEMO"
	cfg := Config{

		Scheme: "http",
		Config: AccessConfig{
			Address: []string{
				"10.1.94.48:8761/eureka",
			},
			Params: map[string]string{
				"username": "test",
				"password": "test",
			},
		},
	}
	e := &eureka{
		id: "1",

		client:   newClient(cfg.getAddress(), cfg.getParams()),
		nodes:    discovery.NewNodesData(),
		services: discovery.NewServices(),
		locker:   sync.RWMutex{},
	}
	app, err := e.GetApp(serviceName)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range app.Nodes() {
		t.Log(node.ID())
	}
	ns, bo := e.nodes.Get(serviceName)
	if bo {
		t.Log(len(ns))
	} else {
		t.Error("nodes error")
	}
}
