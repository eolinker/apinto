package router_http

import (
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/router"
	"net/url"
	"strconv"
	"sync"
)

const (
	routerProfession = "router"
	driverName       = "http"
)

var (
	employeeError = fmt.Errorf("router HTTP employees  not found")
)

func Register() {
	router.RegisterFactory(driverName, newRouterFactory())
}

type routerHttpFactory struct {
}

func (r routerHttpFactory) Create(employeeArr []eosc.IEmployee) (router.IRouterManager, error) {
	return newRouterHttpManager(employeeArr)
}

func newRouterFactory() *routerHttpFactory {
	return &routerHttpFactory{}
}

type routerManager struct {
	servers map[int]router.IRouter
	locker  sync.RWMutex
}

// Create 创建路由管理器
func newRouterHttpManager(employeeArr []eosc.IEmployee) (router.IRouterManager, error) {
	if len(employeeArr) == 0 {
		return nil, employeeError
	}

	configs, err := loadRouterHttpEmployee(employeeArr)
	if err != nil {
		return nil, err
	}

	RM := &routerManager{
		servers: make(map[int]router.IRouter),
		locker:  sync.RWMutex{},
	}

	for port, config := range configs {
		p, _ := strconv.Atoi(port)
		rt := &routerTree{
			listenPort:  p,
			serverState: ServerDown,
			employees:   make(map[string]*httpEmployee),
			targets:     make(map[string]*TargetConfig),
			locker:      sync.RWMutex{},
		}

		rt.tree, err = buildTree(config)
		if err != nil {
			return nil, err
		}
		rt.employees = config
		rt.targets = buildTargetsConfig(config)

		RM.servers[p] = rt
	}

	return RM, nil
}

func (RM *routerManager) Set(port int, newEmployee eosc.IEmployee) error {
	RM.locker.Lock()

	_, has := RM.servers[port]
	if !has {
		rT := &routerTree{
			listenPort:  port,
			serverState: ServerDown,
			employees:   make(map[string]*httpEmployee),
			targets:     make(map[string]*TargetConfig),
			locker:      sync.RWMutex{},
		}
		RM.servers[port] = rT
	}
	RM.locker.Unlock()

	return RM.servers[port].Set(newEmployee)
}

func (RM *routerManager) Delete(port int, id string) error {
	rT, has := RM.servers[port]
	if !has {
		return fmt.Errorf("the port corresponding to the router tree does not exist")
	}
	err := rT.Delete(id)
	if err == NoEmployeeError {
		RM.servers[port] = nil
		return nil
	}
	return err
}

func (RM *routerManager) StartAllServer() {
	for _, server := range RM.servers {
		server.Serve()
	}
}

func (RM *routerManager) ShutDownAllServer() {
	for _, server := range RM.servers {
		server.ShutDown()
	}
}

func (RM *routerManager) StartServer(port int) error {
	rT, has := RM.servers[port]
	if !has {
		return fmt.Errorf("the port corresponding to the router tree does not exist")
	}

	err := rT.Serve()
	return err
}

func (RM *routerManager) ShutDownServer(port int) error {
	rT, has := RM.servers[port]
	if !has {
		return fmt.Errorf("the port corresponding to the router tree does not exist")
	}

	err := rT.ShutDown()
	return err
}

func loadRouterHttpEmployee(employeeArr []eosc.IEmployee) (map[string]map[string]*httpEmployee, error) {
	conf := make(map[string]map[string]*httpEmployee, 0)

	for _, employee := range employeeArr {
		if employee.Driver() != driverName {
			continue
		}
		cData := employee.Config()
		var c router.Config
		err := json.Unmarshal([]byte(cData), &c)
		if err != nil {
			return nil, fmt.Errorf("unmarshal routerEmployee Fail [err]: %s employee data: %s ", err, cData)
		}

		hE := &httpEmployee{
			employeeConfig: cData,
			config:         &c,
		}

		id := c.ID
		if id == "" {
			id = c.Name
		}
		if _, has := conf[c.Listen]; !has {
			conf[c.Listen] = map[string]*httpEmployee{id: hE}
			continue
		}
		conf[c.Listen][id] = hE
	}

	if len(conf) == 0 {
		return nil, fmt.Errorf("router HTTP employees  not found")
	}

	return conf, nil
}

// 构建路由树，并排序匹配的顺序
func buildTree(routerConfigs map[string]*httpEmployee) (router.IRouterHandler, error) {
	tree := make(Tree)

	pathConfigSet := toPathConfig(routerConfigs)
	for _, pc := range pathConfigSet {
		err := tree.Append(pc.pathValue, pc.target)
		if err != nil {
			return nil, err
		}
	}

	return createRouter(tree, RouterPathType), nil
}

func toPathConfig(rcs map[string]*httpEmployee) []*pathConfig {
	pathConfigSet := make([]*pathConfig, 0, len(rcs))

	for _, rc := range rcs {
		config := rc.config
		hosts := config.Host
		if config.Host == nil {
			hosts = []string{"*"}
		}
		for _, host := range hosts {
			for _, rule := range config.Rules {
				location := rule.Location
				if location == "" {
					location = "*/"
				}

				header := "*"
				if rule.Header != nil {
					headerStr, _ := json.Marshal(rule.Header)
					header = string(headerStr)
				}

				query := "*"
				if rule.Query != nil {
					queryStr, _ := json.Marshal(rule.Query)
					query = string(queryStr)
				}

				pathConfigSet = append(pathConfigSet, &pathConfig{
					pathValue: []string{host, location, header, query},
					target:    rule.Target,
				})
			}
		}
	}
	return pathConfigSet
}

func buildTargetsConfig(httpEmployees map[string]*httpEmployee) map[string]*TargetConfig {
	targetConfigSet := make(map[string]*TargetConfig)
	for _, hE := range httpEmployees {
		config := hE.config

		hosts := config.Host
		if config.Host == nil {
			hosts = []string{""}
		}
		for _, host := range hosts {
			for _, rule := range config.Rules {
				location := rule.Location
				header := rule.Header

				query := url.Values{}
				if rule.Query != nil {
					for k, v := range rule.Query {
						query.Add(k, v)
					}
				}

				targetConfigSet[rule.Target] = &TargetConfig{location, host, header, query}
			}
		}
	}
	return targetConfigSet
}
