package consul

import (
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku-eosc/discovery"
	"github.com/hashicorp/consul/api"
	"regexp"
	"strconv"
	"strings"
)

//getConsulClient 创建并返回consul客户端
func getConsulClient(addr string, param map[string]string, scheme string) (*api.Client, error) {
	defaultConfig := api.DefaultConfig()
	//配置信息写入进defaultConfig里
	defaultConfig.Address = addr
	defaultConfig.Scheme = scheme
	if scheme == "https" {
		//TODO
		defaultConfig.TLSConfig = api.TLSConfig{}
	}

	if _, has := param["token"]; has {
		defaultConfig.Token = param["token"]
	}

	client, err := api.NewClient(defaultConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
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
