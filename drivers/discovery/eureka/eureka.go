package eureka

import (
	"context"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc/utils/config"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

type eureka struct {
	drivers.WorkerBase
	client     *client
	nodes      discovery.INodesData
	services   discovery.IServices
	context    context.Context
	cancelFunc context.CancelFunc
	locker     sync.RWMutex
}

//GetApp 获取服务发现中目标服务的app
func (e *eureka) GetApp(serviceName string) (discovery.IApp, error) {
	e.locker.RLock()
	nodes, ok := e.nodes.Get(serviceName)
	e.locker.RUnlock()
	if !ok {
		e.locker.Lock()
		nodes, ok = e.nodes.Get(serviceName)
		if !ok {
			// 开始重新获取
			ns, err := e.client.GetNodeList(serviceName)
			if err != nil {
				e.locker.Unlock()
				return nil, err
			}
			e.nodes.Set(serviceName, ns)
			nodes = ns
		}
		e.locker.Unlock()
	}
	app := discovery.NewApp(nil, e, nil, nodes)
	//将生成的app存入目标服务的app列表
	e.services.Set(serviceName, app.ID(), app)
	return app, nil
}

//Remove 从所有服务app中移除目标app
func (e *eureka) Remove(id string) error {
	e.locker.Lock()
	defer e.locker.Unlock()
	name, count := e.services.Remove(id)
	// 对应services没有app了，移除所有节点
	if count == 0 {
		e.nodes.Del(name)
	}
	return nil
}

//Start 开始服务发现
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
					keys := e.services.AppKeys()
					for _, serviceName := range keys {
						res, err := e.client.GetNodeList(serviceName)
						if err != nil {
							log.Warnf("eureka %s:%w for service %s", e.Name(), discovery.ErrDiscoveryDown, serviceName)
							continue
						}
						//更新目标服务的节点列表
						e.locker.Lock()
						e.nodes.Set(serviceName, res)
						e.locker.Unlock()
						e.services.Update(serviceName, res)
					}
				}
			}

		}
	}()
	return nil
}

//Reset 重置eureka实例配置
func (e *eureka) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	e.client = newClient(cfg.getAddress(), cfg.getParams())
	return nil
}

//Stop 停止服务发现
func (e *eureka) Stop() error {
	e.cancelFunc()
	return nil
}

//CheckSkill 检查目标能力是否存在
func (e *eureka) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}
