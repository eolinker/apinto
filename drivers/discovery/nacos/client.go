package nacos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
)

type client struct {
	address []string
	params  url.Values
}

func newClient(address []string, params url.Values) *client {
	adds := make([]string, 0, len(address))
	for _, a := range address {
		if !strings.HasPrefix(a, "http://") && !strings.HasPrefix(a, "https://") {
			a = fmt.Sprintf("%s://%s", defaultScheme, a)
		}
		adds = append(adds, a)
	}
	return &client{adds, params}
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
			if _, exist := nodes[host.InstanceID]; !exist {
				node := discovery.NewNode(label, host.InstanceID, host.IP, host.Port)
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
	paramsURL := c.params
	paramsURL.Set("serviceName", serviceName)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = paramsURL.Encode()
	response, err := http.DefaultClient.Do(req)
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
