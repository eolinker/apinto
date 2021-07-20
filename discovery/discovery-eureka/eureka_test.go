package discovery_eureka

import (
	"fmt"
	"github.com/eolinker/goku-eosc/discovery"
	"testing"
)

func TestGetApp(t *testing.T) {
	serviceName := "DEMO"
	e := &eureka{
		id:   "1",
		name: "eolinker",
		address: []string{
			"http://39.108.94.48:8761/eureka",
		},
		params: map[string]string{
			"username": "test",
			"password": "test",
		},
		labels:     nil,
		services:   discovery.NewServices(),
		context:    nil,
		cancelFunc: nil,
	}

	app, err := e.GetApp(serviceName)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, node := range app.Nodes() {
		fmt.Println(node.ID())
	}
}
