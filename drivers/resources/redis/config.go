package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Addrs      []string `json:"addrs" label:"redis 节点列表"`
	MasterName string   `json:"master_name" label:"主节点名称"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	DB         int      `json:"db"`
	Mode       string   `json:"mode" enum:"cluster,single" label:"模式"`
	Scopes     []string `json:"scopes" label:"资源组"`
}

func getClient(options *redis.UniversalOptions, mode string) redis.UniversalClient {
	if options.MasterName != "" {
		return redis.NewFailoverClient(options.Failover())
	} else if len(options.Addrs) > 1 || mode == "cluster" {
		return redis.NewClusterClient(options.Cluster())
	}
	simpleClient := redis.NewClient(options.Simple())
	ctx, cf := context.WithTimeout(context.Background(), time.Second)
	defer cf()
	info := simpleClient.Info(ctx, "cluster")
	if info.Err() != nil || !strings.Contains(info.String(), "cluster_enabled:1") {
		return simpleClient
	}

	nodes := simpleClient.ClusterNodes(context.Background())
	if nodes.Err() != nil {
		return simpleClient
	}
	_ = simpleClient.Close()
	nodesContent := nodes.String()
	nodesContent = strings.TrimPrefix(nodesContent, "cluster nodes: ")
	nodesContent = strings.TrimSpace(nodesContent)
	lines := strings.SplitN(nodesContent, "\n", -1)
	nodeAddrs := make([]string, 0, len(lines))
	for _, line := range lines {
		nodeAddrs = append(nodeAddrs, readAddr(line))
	}
	options.Addrs = nodeAddrs
	return redis.NewClusterClient(options.Cluster())
}

func (c *Config) connect() (redis.UniversalClient, error) {
	if len(c.Addrs) == 0 {
		return nil, fmt.Errorf("addrs:%w", eosc.ErrorRequire)
	}
	options := &redis.UniversalOptions{
		Addrs:      c.Addrs,
		MasterName: c.MasterName,
		Username:   c.Username,
		Password:   c.Password,
		DB:         c.DB,
	}
	client := getClient(options, c.Mode)
	if client == nil {
		return nil, fmt.Errorf("get client error")
	}
	err := client.Ping(context.Background()).Err()
	if err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func readAddr(line string) string {
	fields := strings.Fields(line)
	addr := fields[1]

	index := strings.Index(addr, "@")
	if index > 0 {
		addr = addr[:index]
	}
	return addr

}
