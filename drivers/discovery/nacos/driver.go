package nacos

import (
	"github.com/eolinker/apinto/drivers"
	"sync"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

const (
	driverName = "nacos"
)

//Create 创建nacos驱动实例
func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return &nacos{
		WorkerBase: drivers.Worker(id, name),
		client:     newClient(cfg.Config.Address, cfg.getParams()),
		nodes:      discovery.NewNodesData(),
		services:   discovery.NewServices(),
		locker:     sync.RWMutex{},
	}, nil

}
