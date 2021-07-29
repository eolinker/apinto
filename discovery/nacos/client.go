package nacos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/goku-eosc/discovery"
)

type client struct {
	address []string
}

func newClient(address []string, scheme string) *client {
	adds := make([]string, 0, len(address))
	for _, a := range address {
		if !strings.HasPrefix(a, "http://") && !strings.HasPrefix(a, "https://") {
			a = fmt.Sprintf("%s://%s", scheme, a)
		}
		adds = append(adds, a)
	}
	return &client{adds}
}

//GetNodeList 从nacos接入地址中获取对应服务的节点列表
func (c *client) GetNodeList(serviceName string) (discovery.Nodes, error) {
	nodes := make(discovery.Nodes)
	isOk := false
	for _, addr := range c.address {
		ins, err := c.GetInstanceList(addr, serviceName)
		if err != nil {
			log.Info("nacos get node instance list error:", err)
			continue
		}
		isOk = true
		for _, host := range ins.Hosts {
			label := map[string]string{
				"valid":  strconv.FormatBool(host.Valid),
				"marked": strconv.FormatBool(host.Marked),
				"weight": strconv.FormatFloat(host.Weight, 'f', -1, 64),
			}
			if _, exist := nodes[host.InstanceId]; !exist {
				node := discovery.NewNode(label, host.InstanceId, host.Ip, host.Port, "")
				if _, ok := nodes[node.ID()]; ok {
					continue
				}
				nodes[node.ID()] = node
			}
		}
	}
	if !isOk {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodes, nil
}

//GetInstanceList 获取目标地址指定服务名的实例列表
func (c *client) GetInstanceList(addr string, serviceName string) (*Instance, error) {
	addr = addr + instancePath

	response, err := SendRequest(addr, serviceName)
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
