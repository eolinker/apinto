package eureka

import (
	"github.com/eolinker/apinto/drivers"
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestGetApp(t *testing.T) {
	serviceName := "DEMO"
	cfg := Config{
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
		WorkerBase: drivers.Worker("1", name),

		client:   newClient(cfg.getAddress(), cfg.getParams()),
		services: discovery.NewAppContainer(),
		locker:   sync.RWMutex{},
	}
	app, err := e.GetApp(serviceName)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range app.Nodes() {
		t.Log(node.ID())
	}

}
