package consul

import (
	"context"
	"fmt"
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
	services     discovery.IServices
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
							log.Error(err)
							continue
						}

						nodes := make([]discovery.INode, 0, len(nodeSet))
						for k := range nodeSet {
							nodes = append(nodes, nodeSet[k])
						}
						//更新目标服务的节点列表
						c.services.Update(serviceName, nodes)
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

	c.scheme = workerConfig.Scheme
	if c.scheme == "" {
		c.scheme = "http"
	}
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
	return c.services.Remove(id)
}

//GetApp 获取服务发现中目标服务的app
func (c *consul) GetApp(serviceName string) (discovery.IApp, error) {
	nodes := make(map[string]discovery.INode)
	var err error

	oldApp, has := c.services.Get(serviceName)
	if !has {
		nodes, err = c.getNodes(serviceName)
		if err != nil {
			return nil, err
		}
	} else {
		oldAppNodes := oldApp.Nodes()
		for k, node := range oldAppNodes {
			nodes[node.ID()] = oldAppNodes[k]
		}
	}

	app, err := c.Create(serviceName, c.labels, nodes)
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
		for _, node := range clientNodes {
			if _, has := nodeSet[node.ID()]; !has {
				nodeSet[node.ID()] = node
			}
		}
	}

	return nodeSet, nil
}
