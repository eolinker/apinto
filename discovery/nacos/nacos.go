package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
)

const (
	instancePath = "/nacos/v1/ns/instance/list"
)

type nacos struct {
	id         string
	name       string
	address    []string
	params     map[string]string
	labels     map[string]string
	services   discovery.IServices
	context    context.Context
	cancelFunc context.CancelFunc
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
						query := n.getParams(serviceName)
						res, err := n.GetNodeList(query)
						if err != nil {
							continue
						}
						nodes := make([]discovery.INode, 0, len(res))
						for _, v := range res {
							nodes = append(nodes, v)
						}
						//更新目标服务的节点列表
						n.services.Update(serviceName, nodes)
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
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}
	n.address = cfg.Config.Address
	n.params = cfg.Config.Params
	n.labels = cfg.Labels
	return nil
}

//Stop 停止服务发现
func (n *nacos) Stop() error {
	n.cancelFunc()
	return nil
}

//Remove 从所有服务app中移除目标app
func (n *nacos) Remove(id string) error {
	return n.services.Remove(id)
}

//GetApp 获取服务发现中目标服务的app
func (n *nacos) GetApp(serviceName string) (discovery.IApp, error) {
	tmp, ok := n.services.Get(serviceName)
	if !ok {
		var err error
		tmp, err = n.Create(serviceName)
		if err != nil {
			return nil, err
		}
	}
	nodesMap := make(map[string]discovery.INode)
	for _, n := range tmp.Nodes() {
		nodesMap[n.ID()] = n
	}

	app := discovery.NewApp(nil, n, n.getParams(serviceName), nodesMap)
	//将生成的app存入目标服务的app列表
	n.services.Set(serviceName, app.ID(), app)
	return app, nil
}

//Create 创建目标服务的app
func (n *nacos) Create(serviceName string) (discovery.IApp, error) {
	query := n.getParams(serviceName)
	nodes, err := n.GetNodeList(query)
	if err != nil {
		return nil, err
	}
	app := discovery.NewApp(nil, n, query, nodes)
	return app, nil
}

//GetNodeList 从nacos接入地址中获取对应服务的节点列表
func (n *nacos) GetNodeList(query map[string]string) (map[string]discovery.INode, error) {
	nodes := make(map[string]discovery.INode)
	for _, addr := range n.address {
		ins, err := n.GetInstanceList(addr, query)
		if err != nil {
			return nil, err
		}

		for _, host := range ins.Hosts {
			label := map[string]string{
				"valid":  strconv.FormatBool(host.Valid),
				"marked": strconv.FormatBool(host.Marked),
				"weight": strconv.FormatFloat(host.Weight, 'f', -1, 64),
			}
			node := discovery.NewNode(label, host.InstanceId, host.Ip, host.Port, "")
			if _, ok := nodes[node.ID()]; ok {
				continue
			}
			nodes[node.ID()] = node
		}
	}
	return nodes, nil
}

//GetInstanceList 获取目标地址指定服务名的实例列表
func (n *nacos) GetInstanceList(addr string, query map[string]string) (*Instance, error) {
	addr = addr + instancePath
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
		if v, ok := n.labels["schema"]; ok {
			if v == "https" {
				addr = fmt.Sprintf("https://%s", addr)
			}
		}
	}
	response, err := SendRequest(http.MethodGet, addr, query, nil)
	if err != nil {
		return nil, err
	}
	// 解析响应数据
	rawResponseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}
	var instance = &Instance{}
	err = json.Unmarshal(rawResponseData, instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
