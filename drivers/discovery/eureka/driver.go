package eureka

import (
	"github.com/eolinker/apinto/drivers"
	"sync"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

const (
	driverName = "eureka"
)

// Create 创建eureka驱动实例
func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return &eureka{
		WorkerBase: drivers.Worker(id, name),
		client:     newClient(conf.getAddress(), conf.getParams()),
		nodes:      discovery.NewNodesData(),
		services:   discovery.NewServices(),
		locker:     sync.RWMutex{},
	}, nil
}
