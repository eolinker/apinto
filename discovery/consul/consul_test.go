package consul

import (
	"testing"

	"github.com/eolinker/eosc/log"
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
		services:   discovery.NewServices(),
		context:    nil,
		cancelFunc: nil,
	}

	newConsul.Start()

	APP, _ := newConsul.GetApp("consul")

	log.Infof("%s", APP)

	_, _ = newConsul.GetApp("redis")

	select {}
}
