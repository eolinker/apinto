package polaris

import (
	"sync"
	"testing"

	"github.com/eolinker/apinto/discovery"
)

func TestPolarisGetNodes(t *testing.T) {
	accessConfig := &AccessConfig{
		Address:   []string{"127.0.0.1:8091"},
		Namespace: "default",
		Params:    map[string]string{},
	}
	clients := newClients(accessConfig.Address, accessConfig.Namespace, accessConfig.Params)

	newConsul := &polarisDiscovery{
		clients:    clients,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
		context:    nil,
		cancelFunc: nil,
	}

	_ = newConsul.Start()

	APP, _ := newConsul.GetApp("polaris_service")

	t.Log(len(APP.Nodes()))

	select {}
}
