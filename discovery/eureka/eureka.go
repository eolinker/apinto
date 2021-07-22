package eureka

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
)

type eureka struct {
	id         string
	name       string
	address    []string
	params     map[string]string
	labels     map[string]string
	services   discovery.IServices
	context    context.Context
	cancelFunc context.CancelFunc
}

//GetApp 获取服务发现中目标服务的app
func (e *eureka) GetApp(serviceName string) (discovery.IApp, error) {
	app, err := e.Create(serviceName)
	if err != nil {
		return nil, err
	}
	//将生成的app存入目标服务的app列表
	err = e.services.Set(serviceName, app.ID(), app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

//Create 创建目标服务的app
func (e *eureka) Create(serviceName string) (discovery.IApp, error) {
	nodes, err := e.GetNodeList(serviceName)
	if err != nil {
		return nil, err
	}
	attrs := make(discovery.Attrs)
	app := discovery.NewApp(nil, e, attrs, nodes)
	return app, nil
}

//Remove 从所有服务app中移除目标app
func (e *eureka) Remove(id string) error {
	return e.services.Remove(id)
}

//Id 返回 worker id
func (e *eureka) Id() string {
	return e.id
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
						res, err := e.GetNodeList(serviceName)
						if err != nil {
							continue
						}
						nodes := make([]discovery.INode, 0, len(res))
						for _, v := range res {
							nodes = append(nodes, v)
						}
						//更新目标服务的节点列表
						e.services.Update(serviceName, nodes)
					}
				}
			}

		}
	}()
	return nil
}

//Reset 重置eureka实例配置
func (e *eureka) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}
	e.address = cfg.Config.Address
	e.params = cfg.Config.Params
	e.labels = cfg.Labels
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

//GetNodeList 从eureka接入地址中获取对应服务的节点列表
func (e *eureka) GetNodeList(serviceName string) (map[string]discovery.INode, error) {
	nodes := make(map[string]discovery.INode)
	for _, addr := range e.address {
		app, err := e.GetApplication(addr, serviceName)
		if err != nil {
			return nil, err
		}
		for _, ins := range app.Instances {
			if ins.Status != eurekaStatusUp {
				continue
			}
			port := 0
			if ins.Port.Enabled {
				port = ins.Port.Port
			} else if ins.SecurePort.Enabled {
				port = ins.SecurePort.Port
			}
			label := map[string]string{
				"app":      ins.App,
				"hostName": ins.HostName,
			}
			//for k, v := range ins.Metadata {
			//	label[k] = v
			//}
			node := discovery.NewNode(label, ins.InstanceID, ins.IPAddr, port)
			if _, ok := nodes[node.ID()]; ok {
				continue
			}
			nodes[node.ID()] = node
		}
	}
	return nodes, nil
}

//GetApplication 获取每个ip中指定服务名的实例列表
func (e *eureka) GetApplication(addr, serviceName string) (*Application, error) {

	if !strings.Contains(addr, "http://") && !strings.Contains(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
		if v, ok := e.labels["schema"]; ok {
			if v == "https" {
				addr = fmt.Sprintf("https://%s", addr)
			}
		}
	}
	addr = fmt.Sprintf("%s/apps/%s", addr, serviceName)
	res, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, err
	}
	var application = &Application{}
	err = xml.Unmarshal(respBody, application)
	if err != nil {
		return nil, err
	}
	return application, err
}
