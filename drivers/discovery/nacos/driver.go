package nacos

import (
	"fmt"
	"sync"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/discovery"
)

const (
	driverName = "nacos"
)

// Create 创建nacos驱动实例
func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	c, err := newClient(name, cfg.Config.Address, cfg.Config.Params)
	if err != nil {
		return nil, fmt.Errorf("create nacos client fail. err: %w", err)
	}
	return &executor{
		WorkerBase: drivers.Worker(id, name),
		client:     c,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}, nil

}
