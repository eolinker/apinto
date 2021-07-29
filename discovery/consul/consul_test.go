package consul

import (
	"sync"
	"testing"

	"github.com/eolinker/goku-eosc/discovery"
)

func TestConsulGetNodes(t *testing.T) {
	//创建consul

	newConsul := &consul{
		id: "newConsul",
		accessConfig: &AccessConfig{
			Address: []string{"10.1.94.48:8500", "10.1.94.48:8501"},
			Params:  map[string]string{"token": "a92316d8-5c99-4fa0-b4cd-30b9e66718aa"}, //token在10.1.94.48下的/opt/consul/server_config/node_3/conf/acl.hcl文件里
		},
		labels:     map[string]string{"scheme": "http"},
		nodes:      discovery.NewNodesData(),
		services:   discovery.NewServices(),
		locker:     sync.RWMutex{},
		context:    nil,
		cancelFunc: nil,
	}

	newConsul.Start()

	APP, _ := newConsul.GetApp("consul")

	t.Log(APP)

	_, _ = newConsul.GetApp("redis")

	select {}
}
