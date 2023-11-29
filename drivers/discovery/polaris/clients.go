package polaris

import (
	"fmt"

	"github.com/eolinker/eosc/log"
	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/config"

	"github.com/eolinker/apinto/discovery"
)

var defaultNamespace = "default"

type polarisNodeInfo struct {
	id       string
	nodeInfo discovery.NodeInfo
}

func newClients(address []string, namespace string, param map[string]string) polarisClients {
	consumerAPI, err := polaris.NewConsumerAPIByConfig(getConfiguration(address))
	if err != nil {
		log.Warnf("polaris client create fail. addr: %+v  err:%s", address, err)
		return polarisClients{namespace: namespace}
	}
	if len(namespace) == 0 {
		log.Infof("polaris client create not set namespace user default: %s", defaultNamespace)
		namespace = defaultNamespace
	}
	return polarisClients{
		consumerAPI: consumerAPI,
		namespace:   namespace,
	}
}

// getConfiguration 获取北极星配置，完整配置 https://github.com/polarismesh/polaris-go/blob/main/polaris.yaml
func getConfiguration(addresses []string) config.Configuration {
	polarisConfig := config.NewDefaultConfiguration(addresses)
	return polarisConfig
}

// getNodes 通过接入地址获取节点信息
func (c *polarisClients) getNodes(service string) ([]discovery.NodeInfo, error) {
	// 存储节点信息
	nodeList := make([]discovery.NodeInfo, 0, 2)
	// 去重
	nodeIDSet := make(map[string]struct{})
	clientNodes := c.getNodesFromConsumerAPI(c.consumerAPI, service)
	if len(clientNodes) == 0 {
		return nil, discovery.ErrDiscoveryDown
	}
	for _, n := range clientNodes {
		if _, exist := nodeIDSet[n.id]; !exist {
			nodeList = append(nodeList, n.nodeInfo)
		}
		nodeIDSet[n.id] = struct{}{}
	}
	return nodeList, nil
}

// getNodesFromConsumerAPI 从连接的客户端返回健康的节点信息
func (c *polarisClients) getNodesFromConsumerAPI(consumerAPI polaris.ConsumerAPI, service string) []*polarisNodeInfo {
	if consumerAPI == nil {
		log.Warnf("polaris client is nil, get nodes fail.")
		return nil
	}
	req := &polaris.GetInstancesRequest{}
	req.Service = service
	req.Namespace = c.namespace
	instances, err := consumerAPI.GetInstances(req)
	if err != nil {
		return nil
	}
	nodes := make([]*polarisNodeInfo, 0, len(instances.Instances))
	for _, instance := range instances.Instances {
		nodes = append(nodes, &polarisNodeInfo{
			id: fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			nodeInfo: discovery.NodeInfo{
				Ip:     instance.GetHost(),
				Port:   int(instance.GetPort()),
				Labels: instance.GetMetadata(),
			},
		})
	}
	return nodes
}

// Destroy 销毁API，销毁后无法再进行调用
func (c *polarisClients) Destroy() {
	if c.consumerAPI == nil {
		return
	}
	c.consumerAPI.Destroy()
}
