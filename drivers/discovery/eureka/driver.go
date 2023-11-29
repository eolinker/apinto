package eureka

import (
	"sync"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/discovery"
)

const (
	driverName = "eureka"
)

// Create 创建eureka驱动实例
func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return &eureka{
		WorkerBase: drivers.Worker(id, name),
		client:     newClient(conf.getAddress(), conf.getParams()),
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}, nil
}
