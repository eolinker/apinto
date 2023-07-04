package polaris

import (
	"sync"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	driverName = "polaris"
)

// Create 创建北极星驱动实例
func Create(id, name string, workerConfig *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	clients := newClients(workerConfig.Config.Address, workerConfig.Config.Namespace, workerConfig.Config.Params)
	c := &polarisDiscovery{
		WorkerBase: drivers.Worker(id, name),
		clients:    clients,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}
	return c, nil
}
