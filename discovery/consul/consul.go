package consul

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/goku-eosc/discovery"
)

type consul struct {
	id           string
	name         string
	scheme       string
	accessConfig *AccessConfig
	labels       map[string]string
	nodes        discovery.INodesData
	services     discovery.IServices
	locker       sync.RWMutex
	context      context.Context
	cancelFunc   context.CancelFunc
}

//Start 开始服务发现
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
					keys := c.services.AppKeys()
					for _, serviceName := range keys {
						nodeSet, err := c.getNodes(serviceName)
						if err != nil {
							log.Warnf("consul %s:%w for service %s", c.name, discovery.ErrDiscoveryDown, serviceName)
							continue
						}
						//更新目标服务的节点列表
						c.locker.Lock()
						c.nodes.Set(serviceName, nodeSet)
						c.locker.Unlock()
						c.services.Update(serviceName, nodeSet)
					}
				}

			}

		}

	}()
	return nil
}

//Reset 重置consul实例配置
func (c *consul) Reset(config interface{}, workers map[eosc.RequireId]interface{}) error {
	workerConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(config))
	}

	c.accessConfig = &AccessConfig{
		Address: workerConfig.Config.Address,
		Params:  workerConfig.Config.Params,
	}

	c.scheme = workerConfig.getScheme()
	c.labels = workerConfig.Labels

	return nil
}

//Stop 停止服务发现
func (c *consul) Stop() error {
	c.cancelFunc()
	return nil
}

//Remove 从所有服务app中移除目标app
func (c *consul) Remove(id string) error {
	c.locker.Lock()
	defer c.locker.Unlock()
	name, count := c.services.Remove(id)
	if count == 0 {
		c.nodes.Del(name)
	}
	return nil
}

//GetApp 获取服务发现中目标服务的app
func (c *consul) GetApp(serviceName string) (discovery.IApp, error) {
	var err error
	var has bool
	c.locker.RLock()
	nodes, has := c.nodes.Get(serviceName)
	c.locker.RUnlock()
	if !has {
		c.locker.Lock()
		nodes, has = c.nodes.Get(serviceName)
		if !has {
			nodes, err = c.getNodes(serviceName)
			if err != nil {
				c.locker.Unlock()
				return nil, err
			}

			c.nodes.Set(serviceName, nodes)
		}
		c.locker.Unlock()
	}

	app, err := c.Create(serviceName, nil, nodes)
	if err != nil {
		return nil, err
	}
	//将生成的app存入目标服务的app列表
	c.services.Set(serviceName, app.ID(), app)
	return app, nil
}

//Create 创建目标服务的app
func (c *consul) Create(serviceName string, attrs map[string]string, nodes map[string]discovery.INode) (discovery.IApp, error) {
	return discovery.NewApp(nil, c, attrs, nodes), nil
}

//Id 返回 worker id
func (c *consul) Id() string {
	return c.id
}

//CheckSkill 检查目标能力是否存在
func (c *consul) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

//getNodes 通过接入地址获取节点信息
func (c *consul) getNodes(service string) (map[string]discovery.INode, error) {
	nodeSet := make(map[string]discovery.INode)
	ok := false
	for _, addr := range c.accessConfig.Address {
		if !validAddr(addr) {
			log.Errorf("address:%s is invalid", addr)
			continue
		}
		client, err := getConsulClient(addr, c.accessConfig.Params, c.scheme)
		if err != nil {
			log.Error(err)
			continue
		}

		clientNodes := getNodesFromClient(client, service)
		if len(clientNodes) == 0 {
			continue
		}
		ok = true
		for _, node := range clientNodes {
			if _, has := nodeSet[node.ID()]; !has {
				nodeSet[node.ID()] = node
			}
		}
	}
	if !ok {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodeSet, nil
}
