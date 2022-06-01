package nacos

import (
	"context"
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

const (
	instancePath = "/nacos/v1/ns/instance/list"
)

type nacos struct {
	id         string
	name       string
	client     *client
	nodes      discovery.INodesData
	services   discovery.IServices
	context    context.Context
	cancelFunc context.CancelFunc
	locker     sync.RWMutex
}

//Instance nacos 服务实例结构
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

//Id 返回 worker id
func (n *nacos) Id() string {
	return n.id
}

//CheckSkill 检查目标能力是否存在
func (n *nacos) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

//Start 开始服务发现
func (n *nacos) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	n.context = ctx
	n.cancelFunc = cancelFunc
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
					keys := n.services.AppKeys()
					for _, serviceName := range keys {
						res, err := n.client.GetNodeList(serviceName)
						if err != nil {
							log.Warnf("nacos %s:%w for service %s", n.name, discovery.ErrDiscoveryDown, serviceName)
							continue
						}
						//更新目标服务的节点列表
						n.locker.Lock()
						n.nodes.Set(serviceName, res)
						n.locker.Unlock()
						n.services.Update(serviceName, res)

					}
				}
			}

		}
	}()
	return nil
}

//Reset 重置nacos实例配置
func (n *nacos) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	n.client = newClient(cfg.Config.Address, cfg.getParams(), cfg.getScheme())
	return nil
}

//Stop 停止服务发现
func (n *nacos) Stop() error {
	n.cancelFunc()
	return nil
}

//Remove 从所有服务app中移除目标app
func (n *nacos) Remove(id string) error {
	n.locker.Lock()
	defer n.locker.Unlock()
	name, count := n.services.Remove(id)
	if count == 0 {
		n.nodes.Del(name)
	}
	return nil
}

//GetApp 获取服务发现中目标服务的app
func (n *nacos) GetApp(serviceName string) (discovery.IApp, error) {
	n.locker.RLock()
	nodes, ok := n.nodes.Get(serviceName)
	n.locker.RUnlock()
	if !ok {
		n.locker.Lock()
		nodes, ok = n.nodes.Get(serviceName)
		if !ok {
			ns, err := n.client.GetNodeList(serviceName)
			if err != nil {
				n.locker.Unlock()
				return nil, err
			}

			n.nodes.Set(serviceName, ns)
			nodes = ns
		}

		n.locker.Unlock()
	}

	app := discovery.NewApp(nil, n, nil, nodes)
	//将生成的app存入目标服务的app列表
	n.services.Set(serviceName, app.ID(), app)
	return app, nil
}
