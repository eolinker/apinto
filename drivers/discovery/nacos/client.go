package nacos

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
)

type client struct {
	namingClient naming_client.INamingClient
	group        string
	clusters     []string
}

func newClient(name string, address []string, params map[string]string) (*client, error) {
	clientConfig := &constant.ClientConfig{
		LogDir:   fmt.Sprintf("/var/log/apinto/nacos/%s", name),
		CacheDir: fmt.Sprintf("/var/cache/apinto/nacos/%s", name),
		LogLevel: "error",
	}
	//获取namespaceId, username,password
	if v, has := params["namespaceId"]; has {
		clientConfig.NamespaceId = v
	}
	if v, has := params["username"]; has {
		clientConfig.Username = v
	}
	if v, has := params["password"]; has {
		clientConfig.Password = v
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(address))
	var (
		scheme, ipAddr string
		port           uint64
		err            error
	)
	for _, addr := range address {
		schemeIdx := strings.Index(addr, "://")
		if schemeIdx < 0 {
			scheme = defaultScheme
		} else {
			scheme = addr[:schemeIdx]
			addr = addr[schemeIdx+3:]
		}
		portIdx := strings.Index(addr, ":")
		if portIdx < 0 {
			ipAddr = addr
			port = 0
		} else {
			ipAddr = addr[:portIdx]
			portStr := addr[portIdx+1:]
			port, err = strconv.ParseUint(portStr, 10, 64)
			if err != nil {
				return nil, err
			}
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			Scheme: scheme,
			IpAddr: ipAddr,
			Port:   port,
		})
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, err
	}

	c := &client{namingClient: namingClient}
	//获取group, clusters
	if v, has := params["group"]; has {
		c.group = v
	}
	if v, has := params["clusters"]; has {
		clusters := strings.Split(strings.TrimSpace(v), ",")
		c.clusters = clusters
	}

	return c, nil
}

// GetNodeList 从nacos接入地址中获取对应服务的节点列表
func (c *client) GetNodeList(serviceName string) ([]discovery.NodeInfo, error) {
	nodes := make([]discovery.NodeInfo, 0)
	set := make(map[string]struct{})

	for _, addr := range c.address {
		ins, err := c.GetInstanceList(addr, serviceName)
		if err != nil {
			log.Info("nacos get node instance list error:", err)
			continue
		}

		for _, host := range ins.Hosts {
			label := map[string]string{
				"valid":  strconv.FormatBool(host.Valid),
				"marked": strconv.FormatBool(host.Marked),
				"weight": strconv.FormatFloat(host.Weight, 'f', -1, 64),
			}
			if _, exist := set[host.InstanceID]; !exist {
				set[host.InstanceID] = struct{}{}
				nodes = append(nodes, discovery.NodeInfo{
					Ip:     host.IP,
					Port:   host.Port,
					Labels: label,
				})
			}
		}
	}
	if len(nodes) == 0 {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodes, nil
}

// GetInstanceList 获取目标地址指定服务名的实例列表
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
	rawResponseData, err := io.ReadAll(response.Body)
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
