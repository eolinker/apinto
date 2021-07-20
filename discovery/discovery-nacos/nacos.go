package discovery_nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	instancePath = "/nacos/v1/ns/instance/list"
)

type nacos struct {
	id             string
	name			string
	address        []string
	params         map[string]string
	labels		 	map[string]string
	services       discovery.IServices
	context        context.Context
	cancelFunc     context.CancelFunc
}

// return worker id
func (n *nacos) Id() string {
	return n.id
}

// check worker skill
func (n *nacos) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

// start worker
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
					keys := n.services.AppKeys()
					for _, serviceName := range keys {
						query := n.getParams(serviceName)
						res, err := n.GetNodeList(query)
						if err != nil {
							continue
						}
						nodes := make([]discovery.INode, len(res))
						for _, v := range res {
							nodes = append(nodes, v)
						}
						n.services.Update(serviceName, nodes)
					}
				}
			}

		}
	}()
	return nil
}

// update worker config
func (n *nacos) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), eosc.ErrorStructType)
	}
	n.address = cfg.Config.Address
	n.params = cfg.Config.Params
	n.labels = cfg.Labels
	return nil
}

// stop worker
func (n *nacos) Stop() error {
	n.cancelFunc()
	return nil
}

// remove app by app_id
func (n *nacos) Remove(id string) error {
	return n.services.Remove(id)
}

// new app according to serviceName
func (n *nacos) GetApp(serviceName string) (discovery.IApp, error) {
	app, err := n.Create(serviceName)
	if err != nil {
		return nil, err
	}
	n.services.Set(serviceName, app.Id(), app)
	return app, nil
}

// create nacos app
func (n *nacos) Create(serviceName string) (discovery.IApp, error) {
	query := n.getParams(serviceName)
	nodes, err := n.GetNodeList(query)
	if err != nil {
		return nil, err
	}
	app := discovery.NewApp(nil, n, query, nodes)
	return app, nil
}

// get app node list
func (n *nacos) GetNodeList(query map[string]string) (map[string]discovery.INode, error) {
	nodes := make(map[string]discovery.INode)
	for _, addr := range n.address {
		// 获取每个ip中指定服务名的实例列表
		ins, err := n.GetInstanceList(addr, query)
		if err != nil {
			return nil, err
		}

		for _, host := range ins.Hosts {
			label := map[string]string{
				"valid":    strconv.FormatBool(host.Valid),
				"marked":   strconv.FormatBool(host.Marked),
				"weight":   strconv.FormatFloat(host.Weight, 'f', -1, 64),
			}
			node := discovery.NewNode(label, host.InstanceId, host.Ip, host.Port)
			if _, ok := nodes[node.Id()]; ok {
				continue
			}
			nodes[node.Id()] = node
		}
	}
	return nodes, nil
}

// get app instance list
func (n *nacos) GetInstanceList(addr string, query map[string]string) (*Instance, error) {
	addr = addr + instancePath
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
		if v,ok := n.labels["schema"]; ok {
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
