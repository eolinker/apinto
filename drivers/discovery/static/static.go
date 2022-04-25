package static

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	health_check_http "github.com/eolinker/apinto/health-check-http"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
)

var (
	errorStructType = errors.New("error struct type")
)

type static struct {
	id         string
	name       string
	scheme     string
	healthOn   bool
	checker    *health_check_http.HTTPCheck
	context    context.Context
	cancelFunc context.CancelFunc
}

//Id 返回 worker-admin id
func (s *static) Id() string {
	return s.id
}

//Start 开始服务发现
func (s *static) Start() error {
	s.context, s.cancelFunc = context.WithCancel(context.Background())

	return nil
}
func (s *static) reset(cfg *Config) error {

	s.scheme = cfg.getScheme()
	if cfg.Health == nil {
		s.healthOn = false
	} else {
		s.healthOn = cfg.HealthOn
	}
	if s.healthOn {
		if s.checker == nil {
			s.checker = health_check_http.NewHTTPCheck(
				health_check_http.Config{
					Protocol:    cfg.Health.Scheme,
					Method:      cfg.Health.Method,
					URL:         cfg.Health.URL,
					SuccessCode: cfg.Health.SuccessCode,
					Period:      time.Duration(cfg.Health.Period) * time.Second,
					Timeout:     time.Duration(cfg.Health.Timeout) * time.Millisecond,
				})
		} else {
			s.checker.Reset(
				health_check_http.Config{
					Protocol:    cfg.Health.Scheme,
					Method:      cfg.Health.Method,
					URL:         cfg.Health.URL,
					SuccessCode: cfg.Health.SuccessCode,
					Period:      time.Duration(cfg.Health.Period) * time.Second,
					Timeout:     time.Duration(cfg.Health.Timeout) * time.Millisecond,
				},
			)
		}
	} else {
		if s.checker != nil {
			s.checker.Stop()
			s.checker = nil
		}
	}
	return nil
}

//Reset 重置静态服务发现实例配置
func (s *static) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), errorStructType)
	}
	return s.reset(cfg)
}

//Stop 停止服务发现
func (s *static) Stop() error {

	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *static) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

//GetApp 获取服务发现中目标服务的app
func (s *static) GetApp(config string) (discovery.IApp, error) {
	app, err := s.decode(config)
	if err != nil {
		return nil, err
	}

	return app, nil
}

//Remove 从所有服务app中移除目标app
func (s *static) Remove(id string) error {
	return nil
}

//Node 静态服务发现的节点类型
type Node struct {
	labels map[string]string
	ip     string
	port   int
}

//decode 通过配置生成app
func (s *static) decode(config string) (discovery.IApp, error) {
	words := fields(config)

	nodes := make(map[string]discovery.INode)

	index := 0
	var node *Node
	attrs := make(discovery.Attrs)
	for _, word := range words {

		if word == ";" {
			if node != nil {
				scheme, has := node.labels["scheme"]
				if !has {
					scheme = s.scheme
				}
				n := discovery.NewNode(node.labels, fmt.Sprintf("%s:%d", node.ip, node.port), node.ip, node.port, scheme)
				nodes[n.ID()] = n
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
		scheme, has := node.labels["scheme"]
		if !has {
			scheme = s.scheme
		}
		n := discovery.NewNode(node.labels, fmt.Sprintf("%s:%d", node.ip, node.port), node.ip, node.port, scheme)
		nodes[n.ID()] = n
	}
	index = 0
	node = nil

	agent := (discovery.IHealthChecker)(nil)
	if s.checker != nil {
		agent, _ = s.checker.Agent()
	}

	app := discovery.NewApp(agent, s, attrs, nodes)
	return app, nil
}

func fields(str string) []string {
	words := strings.FieldsFunc(strings.Join(strings.Split(str, ";"), " ; "), func(r rune) bool {
		return unicode.IsSpace(r)
	})
	return words
}

//validIP 判断ip是否合法
func validIP(ip string) bool {
	match, err := regexp.MatchString(`^(?:(?:1[0-9][0-9]\.)|(?:2[0-4][0-9]\.)|(?:25[0-5]\.)|(?:[1-9][0-9]\.)|(?:[0-9]\.)){3}(?:(?:1[0-9][0-9])|(?:2[0-4][0-9])|(?:25[0-5])|(?:[1-9][0-9])|(?:[0-9]))$`, ip)
	if err != nil {
		return false
	}
	return match
}
