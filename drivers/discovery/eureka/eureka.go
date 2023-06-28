package eureka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc/utils/config"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

var _ discovery.IDiscovery = (*eureka)(nil)

type eureka struct {
	drivers.WorkerBase
	client     *client
	services   discovery.IAppContainer
	context    context.Context
	cancelFunc context.CancelFunc
	locker     sync.RWMutex
}

// GetApp 获取服务发现中目标服务的app
func (e *eureka) GetApp(serviceName string) (discovery.IApp, error) {
	e.locker.RLock()
	app, ok := e.services.GetApp(serviceName)
	e.locker.RUnlock()
	if ok {
		return app.Agent(), nil
	}

	e.locker.Lock()
	app, ok = e.services.GetApp(serviceName)
	if ok {
		return app.Agent(), nil
	}

	// 开始重新获取
	ns, err := e.client.GetNodeList(serviceName)
	if err != nil {
		log.Errorf("%s get %s node list error: %v", driverName, serviceName, err)
	}
	app = e.services.Set(serviceName, ns)
	e.locker.Unlock()

	return app.Agent(), nil
}

// Start 开始服务发现
func (e *eureka) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	e.context = ctx
	e.cancelFunc = cancelFunc
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
					keys := e.services.Keys()
					for _, serviceName := range keys {
						res, err := e.client.GetNodeList(serviceName)
						if err != nil {
							log.Warnf("eureka %s:%w for service %s", e.Name(), discovery.ErrDiscoveryDown, serviceName)
						}
						//更新目标服务的节点列表
						e.services.Set(serviceName, res)
					}
				}
			}

		}
	}()
	return nil
}

// Reset 重置eureka实例配置
func (e *eureka) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	e.client = newClient(cfg.getAddress(), cfg.getParams())
	return nil
}

// Stop 停止服务发现
func (e *eureka) Stop() error {
	e.cancelFunc()
	return nil
}

// CheckSkill 检查目标能力是否存在
func (e *eureka) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}
