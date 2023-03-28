package consul

import (
	"strings"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc/log"
	"github.com/hashicorp/consul/api"
)

type consulNodeInfo struct {
	id       string
	nodeInfo discovery.NodeInfo
}

func newClients(addrs []string, param map[string]string) *consulClients {
	clients := make([]*api.Client, 0, len(addrs))

	defaultConfig := api.DefaultConfig()
	if _, has := param["token"]; has {
		defaultConfig.Token = param["token"]
	}
	if _, has := param["namespace"]; has {
		defaultConfig.Namespace = param["namespace"]
	}

	for _, addr := range addrs {
		//解析addr, client配置需要区分scheme和host
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			defaultConfig.Scheme = defaultScheme
			defaultConfig.Address = addr
		} else {
			idx := strings.Index(addr, "://")
			defaultConfig.Scheme = addr[:idx]
			defaultConfig.Address = addr[idx+3:]
		}

		client, err := api.NewClient(defaultConfig)
		if err != nil {
			log.Warnf("consul client create fail. addr: %s  err:%s", addr, err)
			continue
		}

		clients = append(clients, client)
	}

	return &consulClients{clients: clients}
}

// getNodes 通过接入地址获取节点信息
func (c *consulClients) getNodes(service string) ([]discovery.NodeInfo, error) {
	nodeList := make([]discovery.NodeInfo, 0, 2)
	nodeIDSet := make(map[string]struct{})
	for _, client := range c.clients {
		clientNodes := getNodesFromClient(client, service)
		if len(clientNodes) == 0 {
			continue
		}
		for _, n := range clientNodes {
			if _, exist := nodeIDSet[n.id]; !exist {
				nodeList = append(nodeList, n.nodeInfo)
			}
			nodeIDSet[n.id] = struct{}{}
		}
	}

	return nodeList, discovery.ErrDiscoveryDown
}

// getNodesFromClient 从连接的客户端返回健康的节点信息
func getNodesFromClient(client *api.Client, service string) []*consulNodeInfo {
	queryOptions := &api.QueryOptions{}
	serviceEntryArr, _, err := client.Health().Service(service, "", true, queryOptions)
	if err != nil {
		return nil
	}

	nodes := make([]*consulNodeInfo, 0, len(serviceEntryArr))
	for _, serviceEntry := range serviceEntryArr {
		nodes = append(nodes, &consulNodeInfo{
			id: serviceEntry.Service.ID,
			nodeInfo: discovery.NodeInfo{
				Ip:     serviceEntry.Service.Address,
				Port:   serviceEntry.Service.Port,
				Labels: serviceEntry.Service.Meta,
			},
		})
	}

	return nodes
}
