package consul

import (
	"sync"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers"
)

const (
	driverName = "consul"
)

// Create 创建consul驱动实例
func Create(id, name string, workerConfig *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	clients := newClients(workerConfig.Config.Address, workerConfig.Config.Params)

	c := &consul{
		WorkerBase: drivers.Worker(id, name),
		clients:    clients,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}
	return c, nil
}
