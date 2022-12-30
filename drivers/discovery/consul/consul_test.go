package consul

import (
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestConsulGetNodes(t *testing.T) {
	//创建consul
	accessConfig := &AccessConfig{
		Address: []string{"http://172.18.166.219:8500"},
		//Params:  map[string]string{"token": "a92316d8-5c99-4fa0-b4cd-30b9e66718aa"}, //token在10.1.94.48下的/opt/consul/server_config/node_3/conf/acl.hcl文件里
	}
	clients := newClients(accessConfig.Address, accessConfig.Params)

	newConsul := &consul{
		clients:    clients,
		nodes:      discovery.NewNodesData(),
		services:   discovery.NewServices(),
		locker:     sync.RWMutex{},
		context:    nil,
		cancelFunc: nil,
	}

	newConsul.Start()

	APP, _ := newConsul.GetApp("apinto-test")

	t.Log(APP)

	_, _ = newConsul.GetApp("redis")

	select {}
}
