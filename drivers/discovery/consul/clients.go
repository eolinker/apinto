package consul

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/discovery"
	"github.com/hashicorp/consul/api"
)

func newClients(addrs []string, param map[string]string, scheme string) (*consulClients, error) {
	clients := make([]*api.Client, 0, len(addrs))

	defaultConfig := api.DefaultConfig()
	defaultConfig.Scheme = scheme
	if _, has := param["token"]; has {
		defaultConfig.Token = param["token"]
	}
	if _, has := param["namespace"]; has {
		defaultConfig.Namespace = param["namespace"]
	}

	hasClientFlag := false
	for _, addr := range addrs {
		if !validAddr(addr) {
			log.Warnf("consul address:%s is invalid", addr)
			continue
		}

		defaultConfig.Address = addr
		client, err := api.NewClient(defaultConfig)
		if err != nil {
			log.Warnf("consul client create fail. addr: %s  err:%s", addr, err)
			continue
		}
		hasClientFlag = true
		clients = append(clients, client)
	}

	if !hasClientFlag {
		return nil, fmt.Errorf("consul create clients fail")
	}

	return &consulClients{clients: clients}, nil
}

//getNodes 通过接入地址获取节点信息
func (c *consulClients) getNodes(service string) (map[string]discovery.INode, error) {
	nodeSet := make(map[string]discovery.INode)
	ok := false
	for _, client := range c.clients {
		clientNodes := getNodesFromClient(client, service)
		if len(clientNodes) == 0 {
			continue
		}
		ok = true
		for _, node := range clientNodes {
			if _, has := nodeSet[node.ID()]; !has {
				nodeSet[node.ID()] = node
			}
		}
	}
	if !ok {
		return nil, discovery.ErrDiscoveryDown
	}
	return nodeSet, nil
}

//getNodesFromClient 从连接的客户端返回健康的节点信息
func getNodesFromClient(client *api.Client, service string) []discovery.INode {
	queryOptions := &api.QueryOptions{}
	serviceEntryArr, _, err := client.Health().Service(service, "", true, queryOptions)
	if err != nil {
		return nil
	}

	nodes := make([]discovery.INode, 0, len(serviceEntryArr))
	for _, serviceEntry := range serviceEntryArr {
		nodeAddr := serviceEntry.Node.Address
		addrSlide := append(strings.Split(nodeAddr, ":"))
		ip := addrSlide[0]
		var port int
		if len(addrSlide) > 1 {
			port, err = strconv.Atoi(addrSlide[1])
			if err != nil {
				log.Error(err)
				continue
			}
		}

		newNode := discovery.NewNode(serviceEntry.Service.Meta, serviceEntry.Node.ID, ip, port, "")
		nodes = append(nodes, newNode)
	}

	return nodes
}

//validAddr 判断地址是否合法
func validAddr(addr string) bool {
	c := strings.Split(addr, ":")
	if len(c) < 2 {
		return false
	}
	ip := c[0]
	if !validIP(ip) {
		return false
	}
	_, err := strconv.Atoi(c[1])
	if err != nil {
		return false
	}

	return true
}

//validIP 判断ip是否合法
func validIP(ip string) bool {
	match, err := regexp.MatchString(`^(?:(?:1[0-9][0-9]\.)|(?:2[0-4][0-9]\.)|(?:25[0-5]\.)|(?:[1-9][0-9]\.)|(?:[0-9]\.)){3}(?:(?:1[0-9][0-9])|(?:2[0-4][0-9])|(?:25[0-5])|(?:[1-9][0-9])|(?:[0-9]))$`, ip)
	if err != nil {
		return false
	}
	return match
}
