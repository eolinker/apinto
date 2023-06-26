package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"strconv"
	"strings"

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

// GetNodeList 从nacosClient获取对应服务的节点列表
func (c *client) GetNodeList(serviceName string) ([]discovery.NodeInfo, error) {
	nodes := make([]discovery.NodeInfo, 0)
	set := make(map[string]struct{})

	instances, err := c.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		Clusters:    c.clusters,
		GroupName:   c.group,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}

	for _, ins := range instances {
		label := map[string]string{
			"weight": strconv.FormatFloat(ins.Weight, 'f', -1, 64),
		}
		//ins的instanceID可能为空
		instanceID := fmt.Sprintf("%s:%d", ins.Ip, ins.Port)
		if _, exist := set[instanceID]; !exist {
			set[instanceID] = struct{}{}

			for k, v := range ins.Metadata {
				label[k] = v
			}
			nodes = append(nodes, discovery.NodeInfo{
				Ip:     ins.Ip,
				Port:   int(ins.Port),
				Labels: label,
			})
		}
	}

	if len(nodes) == 0 {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodes, nil
}
