package polaris

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
)

var _ discovery.IDiscovery = (*polarisDiscovery)(nil)
var _ eosc.IWorker = (*polarisDiscovery)(nil)

type polarisDiscovery struct {
	drivers.WorkerBase
	clients    polarisClients
	services   discovery.IAppContainer
	locker     sync.RWMutex
	context    context.Context
	cancelFunc context.CancelFunc
}

func (p *polarisDiscovery) GetApp(serviceName string) (discovery.IApp, error) {
	var err error
	var has bool
	p.locker.RLock()
	app, has := p.services.GetApp(serviceName)
	p.locker.RUnlock()
	if has {
		return app.Agent(), nil
	}

	p.locker.Lock()
	defer p.locker.Unlock()
	app, has = p.services.GetApp(serviceName)
	if has {
		return app.Agent(), nil
	}

	nodes, err := p.clients.getNodes(serviceName)
	if err != nil {
		log.Errorf("%s get %s node list error: %v", driverName, serviceName, err)
	}
	app = p.services.Set(serviceName, nodes)
	return app.Agent(), nil
}

func (p *polarisDiscovery) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.context = ctx
	p.cancelFunc = cancelFunc

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
					// 获取现有服务app的服务名名称列表，并从注册中心获取目标服务名的节点列表
					keys := p.services.Keys()
					for _, serviceName := range keys {
						nodeSet, err := p.clients.getNodes(serviceName)
						if err != nil {
							log.Warnf("polaris %s:%s for service %s", p.Name(), discovery.ErrDiscoveryDown, serviceName)
						}
						// 更新目标服务的节点列表
						p.services.Set(serviceName, nodeSet)
					}
				}
			}
		}
	}()
	return nil
}

func (p *polarisDiscovery) Reset(cfg interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	workerConfig, ok := cfg.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(cfg))
	}
	oldClients := p.clients

	clients := newClients(workerConfig.Config.Address, workerConfig.Config.Namespace, workerConfig.Config.Params)
	p.clients = clients
	// 销毁老的api
	oldClients.Destroy()
	return nil
}

func (p *polarisDiscovery) Stop() error {
	p.cancelFunc()
	// 销毁老的api
	p.clients.Destroy()
	return nil
}

func (p *polarisDiscovery) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}
