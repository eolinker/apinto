package static

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/config"
	"github.com/google/uuid"
	"reflect"
	"strconv"
	"strings"
)

var (
	errorStructType = errors.New("error struct type")
)

type static struct {
	drivers.WorkerBase
	handler   *HeathCheckHandler
	isRunning bool
	cfg       *Config
	services  discovery.IAppContainer
}

// Start 开始服务发现
func (s *static) Start() error {

	handler := s.handler
	s.isRunning = true
	if handler != nil {
		return nil
	}
	handler = NewHeathCheckHandler(s.services, s.cfg)

	return nil
}

// Reset 重置静态服务发现实例配置
func (s *static) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf), errorStructType)
	}

	if reflect.DeepEqual(cfg, s.cfg) {
		return nil
	}
	s.cfg = cfg

	if s.isRunning {
		ck := s.handler
		if ck != nil {
			return ck.reset(cfg)
		}
		return s.Start()
	}
	return nil
}

// Stop 停止服务发现
func (s *static) Stop() error {
	s.isRunning = false
	handler := s.handler
	if handler == nil {
		return nil
	}

	handler.stop()
	return nil
}

// CheckSkill 检查目标能力是否存在
func (s *static) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

// GetApp 获取服务发现中目标服务的app
func (s *static) GetApp(config string) (discovery.IApp, error) {

	app, err := s.decode(config)
	if err != nil {
		return nil, err
	}

	return app.Agent(), nil
}

// Remove 从所有服务app中移除目标app
func (s *static) Remove(id string) error {
	return nil
}

// Node 静态服务发现的节点类型
type Node struct {
	labels map[string]string
	ip     string
	port   int
}

// decode 通过配置生成app
func (s *static) decode(config string) (discovery.IAppAgent, error) {
	words := fields(config)

	nodes := make([]discovery.NodeInfo, 0, 5)

	index := 0
	var node *Node
	for _, word := range words {

		if word == ";" {
			if node != nil {
				nodes = append(nodes, discovery.NodeInfo{
					Ip:     node.ip,
					Port:   node.port,
					Labels: node.labels,
				})
			}
			index = 0
			node = nil
			continue
		}

		switch index {
		case 0:
			{
				// 域名+端口
				node = &Node{
					labels: map[string]string{},
					ip:     "",
					port:   0,
				}
				vs := strings.Split(word, ":")
				// 先判断是否是IP端口模式
				if !validIP(vs[0]) {
					if strings.Contains(vs[0], "=") {
						// 计入全局的属性
						args := strings.Split(vs[0], "=")
						if len(args) > 1 {
							node.labels[args[0]] = args[1]
						}
						break
					}

				}
				if len(vs) > 2 {
					return nil, fmt.Errorf("decode ip:port failt for[%s]", word)
				}
				node.ip = vs[0]
				if len(vs) == 2 {
					port, _ := strconv.Atoi(vs[1])
					node.port = port
				}

			}
		default:
			{
				// label集合
				args := strings.Split(word, "=")
				if len(args) > 1 {
					node.labels[args[0]] = args[1]
				}
			}
		}
		index++
	}
	if node != nil {
		nodes = append(nodes, discovery.NodeInfo{
			Ip:     node.ip,
			Port:   node.port,
			Labels: node.labels,
		})
	}
	index = 0
	node = nil

	app := s.services.Set(uuid.New().String(), nodes)
	return app, nil
}
