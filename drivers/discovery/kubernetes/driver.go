package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

const (
	driverName = "kubernetes"
)

// Create 创建nacos驱动实例
func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if cfg == nil || cfg.Config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	c, err := newClient(ctx, name, cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("create nacos client fail. err: %w", err)
	}
	return &executor{
		WorkerBase: drivers.Worker(id, name),
		ctx:        ctx,
		cancelFunc: cancelFunc,
		client:     c,
		services:   discovery.NewAppContainer(),
		locker:     sync.RWMutex{},
	}, nil

}
