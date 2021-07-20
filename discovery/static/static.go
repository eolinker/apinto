package static

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	health_check_http "github.com/eolinker/goku-eosc/health-check-http"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
)

const name = "static"

type static struct {
	id         string
	name       string
	labels     map[string]string
	apps       map[string]discovery.IApp
	locker     sync.RWMutex
	healthOn   bool
	checker    *health_check_http.HttpCheck
	context    context.Context
	cancelFunc context.CancelFunc
}

//Id 返回 worker id
func (s *static) Id() string {
	return s.id
}

//Start 开始服务发现
func (s *static) Start() error {
	s.context, s.cancelFunc = context.WithCancel(context.Background())

	return nil
}

//Reset 重置静态服务发现实例配置
func (s *static) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), eosc.ErrorStructType)
	}
	s.locker.Lock()
	s.labels = cfg.Labels
	s.locker.Unlock()
	if cfg.Health == nil {
		s.healthOn = false
	} else {
		s.healthOn = cfg.HealthOn
	}
	if s.healthOn {
		if s.checker == nil {
			s.checker = health_check_http.NewHttpCheck(
				health_check_http.Config{
					Protocol:    cfg.Health.Protocol,
					Method:      cfg.Health.Method,
					Url:         cfg.Health.URL,
					SuccessCode: cfg.Health.SuccessCode,
					Period:      time.Duration(cfg.Health.Period) * time.Second,
					Timeout:     time.Duration(cfg.Health.Timeout) * time.Millisecond,
				})
		} else {
			s.checker.Reset(
				health_check_http.Config{
					Protocol:    cfg.Health.Protocol,
					Method:      cfg.Health.Method,
					Url:         cfg.Health.URL,
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

//Stop 停止服务发现
func (s *static) Stop() error {
	for _, a := range s.apps {
		a.Close()
	}
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
	s.locker.Lock()
	s.apps[app.ID()] = app
	s.locker.Unlock()
	return app, nil
}

//Remove 从所有服务app中移除目标app
func (s *static) Remove(id string) error {
	s.locker.Lock()
	delete(s.apps, id)
	s.locker.Unlock()
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
			index = 0
			node = nil
			continue
		}
		l := len(word)
		value := word
		if word[l-1] == ';' {
			value = word[:l-1]
		}
		switch index {
		case 0:
			{
				// 域名+端口
				node = new(Node)
				vs := strings.Split(value, ":")
				// 先判断是否是IP端口模式
				if !validIP(vs[0]) {
					// 若不是IP端口模式，则计入全局的属性
					args := strings.Split(vs[0], "=")
					if len(args) > 1 {
						node.labels[args[0]] = args[1]
					}
					break
				}
				if len(vs) > 2 {
					return nil, fmt.Errorf("decode ip:port failt for[%s]", value)
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
				args := strings.Split(value, "=")
				if len(args) > 1 {
					node.labels[args[0]] = args[1]
				}
			}
		}

		if word[l-1] == ';' {
			n := discovery.NewNode(node.labels, fmt.Sprintf("%s:%d", node.ip, node.port), node.ip, node.port)
			nodes[n.ID()] = n
			index = 0
			node = nil
		} else {
			index++
		}
	}

	//若开启了健康检查，则为app分配一个健康检查agent
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
