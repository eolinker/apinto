package consul

import (
	"context"
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers"
	"sync"
	"time"

	"github.com/eolinker/eosc/utils/config"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

type consul struct {
	drivers.WorkerBase
	clients    *consulClients
	services   discovery.IAppContainer
	locker     sync.RWMutex
	context    context.Context
	cancelFunc context.CancelFunc
}

// Start 开始服务发现
func (c *consul) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	c.context = ctx
	c.cancelFunc = cancelFunc

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
	EXIT:
		for {
			select {
			case <-ctx.Done():
				break EXIT
			case <-ticker.C:
				{
					//获取现有服务app的服务名名称列表，并从注册中心获取目标服务名的节点列表
					keys := c.services.Keys()
					for _, serviceName := range keys {
						nodeSet, err := c.clients.getNodes(serviceName)
						if err != nil {
							log.Warnf("consul %s:%s for service %s", c.Name(), discovery.ErrDiscoveryDown, serviceName)
							continue
						}
						//更新目标服务的节点列表
						c.services.Set(serviceName, nodeSet)
					}
				}

			}

		}

	}()
	return nil
}

// Reset 重置consul实例配置
func (c *consul) Reset(cfg interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	workerConfig, ok := cfg.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(cfg))
	}

	clients := newClients(workerConfig.Config.Address, workerConfig.Config.Params)

	c.clients = clients
	return nil
}

// Stop 停止服务发现
func (c *consul) Stop() error {
	c.cancelFunc()
	return nil
}

// GetApp 获取服务发现中目标服务的app
func (c *consul) GetApp(serviceName string) (discovery.IAppAgent, error) {
	var err error
	var has bool
	c.locker.RLock()
	app, has := c.services.GetApp(serviceName)
	c.locker.RUnlock()
	if has {
		return app, nil
	}

	c.locker.Lock()
	defer c.locker.Unlock()
	app, has = c.services.GetApp(serviceName)
	if has {
		return app, nil
	}

	nodes, err := c.clients.getNodes(serviceName)
	if err != nil {
		log.Errorf("%s get %s node list error: %v", driverName, serviceName, err)
	}
	app = c.services.Set(serviceName, nodes)

	return app, nil

}

// CheckSkill 检查目标能力是否存在
func (c *consul) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}
