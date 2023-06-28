package nacos

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"sync"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
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
	return &nacos{
		WorkerBase: drivers.Worker(id, name),
		client:     c,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}, nil

}
