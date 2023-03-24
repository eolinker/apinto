package eureka

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc/log"
)

type client struct {
	address []string
	params  url.Values
}

func newClient(address []string, params url.Values) *client {
	return &client{address, params}
}

// GetNodeList 从eureka接入地址中获取对应服务的节点列表
func (c *client) GetNodeList(serviceName string) ([]discovery.NodeInfo, error) {
	isOk := false
	nodes := make([]discovery.NodeInfo, 0, 5)
	sets := make(map[string]struct{})
	for _, addr := range c.address {
		app, err := c.GetApplication(addr, serviceName)
		if err != nil {
			log.Info("eureka get node instance list error:", err)
			continue
		}
		isOk = true
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

			if _, exist := sets[ins.InstanceID]; !exist {
				node := discovery.NodeInfo{
					Ip:   ins.IPAddr,
					Port: port,
					Labels: map[string]string{
						"app":      ins.App,
						"hostName": ins.HostName,
					},
					Scheme: ins.Status,
				}
				nodes = append(nodes, node)
			}

		}
	}
	if !isOk {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodes, nil
}

// GetApplication 获取每个ip中指定服务名的实例列表
func (c *client) GetApplication(addr string, serviceName string) (*Application, error) {
	addr = fmt.Sprintf("%s/apps/%s", addr, serviceName)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = c.params.Encode()
	res, err := http.DefaultClient.Do(req)
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
		return nil, fmt.Errorf("http Status:%d", res.StatusCode)
	}
	var application = &Application{}
	err = xml.Unmarshal(respBody, application)
	if err != nil {
		return nil, err
	}
	return application, err
}
