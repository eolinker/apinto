package nacos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/apinto/discovery"

	"github.com/eolinker/eosc/utils/config"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
)

var _ discovery.IDiscovery = (*executor)(nil)
var _ eosc.IWorkerDestroy = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	client     *client
	services   discovery.IAppContainer
	context    context.Context
	cancelFunc context.CancelFunc
	locker     sync.RWMutex
}

func (n *executor) Destroy() error {

	if n.client != nil {
		n.client.namingClient.CloseClient()
		n.client = nil
	}
	return nil
}

// GetApp 获取服务发现中目标服务的app
func (n *executor) GetApp(serviceName string) (discovery.IApp, error) {
	n.locker.RLock()
	app, ok := n.services.GetApp(serviceName)
	n.locker.RUnlock()
	if ok {
		return app.Agent(), nil
	}

	n.locker.Lock()
	defer n.locker.Unlock()
	app, ok = n.services.GetApp(serviceName)
	if ok {
		return app.Agent(), nil
	}

	ns, err := n.client.GetNodeList(serviceName)
	if err != nil {
		log.Warnf("%s get %s node list error: %v", driverName, serviceName, err)
	}

	app = n.services.Set(serviceName, ns)

	return app.Agent(), nil
}

// Instance executor 服务实例结构
type Instance struct {
	Hosts []struct {
		Valid      bool    `json:"valid"`
		Marked     bool    `json:"marked"`
		InstanceID string  `json:"instanceId"`
		Port       int     `json:"port"`
		IP         string  `json:"ip"`
		Weight     float64 `json:"weight"`
	}
}

// CheckSkill 检查目标能力是否存在
func (n *executor) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

// Start 开始服务发现
func (n *executor) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	n.context = ctx
	n.cancelFunc = cancelFunc
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
	EXIT:
		for {
			log.Debug("continue go")
			select {
			case <-ctx.Done():

				break EXIT
			case <-ticker.C:
				{
					log.Debug("continue ticket")
					//获取现有服务app的服务名名称列表，并从注册中心获取目标服务名的节点列表
					keys := n.services.Keys()
					for _, serviceName := range keys {
						res, err := n.client.GetNodeList(serviceName)
						if err != nil {
							log.Warnf("executor %s:%w for service %s", n.Name(), discovery.ErrDiscoveryDown, serviceName)
						}
						//更新目标服务的节点列表
						n.services.Set(serviceName, res)
					}
				}
			}

		}
	}()
	return nil
}

// Reset 重置nacos实例配置
func (n *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	nClient, err := newClient(n.Name(), cfg.Config.Address, cfg.Config.Params)
	if err != nil {
		return fmt.Errorf("create executor client fail. err: %w", err)
	}
	if n.client != nil {
		n.client.namingClient.CloseClient()
		n.client = nil
	}
	n.client = nClient
	return nil
}

// Stop 停止服务发现
func (n *executor) Stop() error {
	n.cancelFunc()
	return n.Destroy()
}
