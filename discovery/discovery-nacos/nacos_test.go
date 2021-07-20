package discovery_nacos

import (
	"fmt"
	"github.com/eolinker/goku-eosc/discovery"
	"testing"
)

func TestNacos_GetApp(t *testing.T) {
	serviceName := "nacos.naming.serviceName"
	n := &nacos{
		address: []string{
			"39.108.94.48:8848",
		},
		params: map[string]string{
			"username": "test",
			"password": "test",
		},
		services:   discovery.NewServices(),
		context:    nil,
		cancelFunc: nil,
	}
	app, _ := n.GetApp(serviceName)
	for _, node := range app.Nodes() {
		fmt.Println(node.ID())
	}
}
