package router_http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/listener"
	"github.com/eolinker/eosc/log"
	listener_manager "github.com/eolinker/goku-eosc/listener-manager"
	"github.com/eolinker/goku-eosc/router"
	"net/http"
	"sync"
)

const (
	ServerUp   = "UP"
	ServerDown = "DOWN"
)

var NoEmployeeError = errors.New("no employee exist")

type routerTree struct {
	listenPort  int
	server      *http.Server
	serverState string
	tree        router.IRouterHandler
	employees   map[string]*httpEmployee // key为实例id
	targets     map[string]*TargetConfig // key为target
	locker      sync.RWMutex
}

type httpEmployee struct {
	employeeConfig string
	config         *router.Config
}

type pathConfig struct {
	pathValue []string
	target    string
}

func (r *routerTree) Set(newEmployee eosc.IWorker) error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if newEmployee.Driver() != driverName {
		return fmt.Errorf("router set employee fail. the employee's driver isn't HTTP ")
	}

	id := newEmployee.ID()
	if id == "" {
		id = newEmployee.Name()
	}

	// 若更新的实例在原有配置中已存在且配置相同，则直接返回
	newEmployeeConfig := newEmployee.Config()
	if oldHE, has := r.employees[id]; has && oldHE.employeeConfig == newEmployeeConfig {
		return nil
	}

	TempEmployees := cloneEmployees(r.employees)
	var c router.Config
	err := json.Unmarshal([]byte(newEmployeeConfig), &c)
	if err != nil {
		return fmt.Errorf("unmarshal routerEmployee Fail [err]: %s employee data: %s ", err, newEmployeeConfig)
	}
	TempEmployees[id] = &httpEmployee{newEmployeeConfig, &c}

	NewTree, err := buildTree(TempEmployees)
	if err != nil {
		return fmt.Errorf("set employee fail err: %s", err)
	}

	r.tree = NewTree
	r.employees = TempEmployees
	r.targets = buildTargetsConfig(r.employees)

	return nil
}

func (r *routerTree) Delete(id string) error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, has := r.employees[id]; !has {
		return fmt.Errorf("delete employee fail. Employee id %s not exist", id)
	}

	TempEmployees := cloneEmployees(r.employees)
	delete(TempEmployees, id)
	if len(TempEmployees) == 0 {
		// 若该端口下已没有路由实例，则关闭Server 并在路由管理器中删除本路由树
		r.ShutDown()
		return NoEmployeeError
	}

	NewTree, err := buildTree(TempEmployees)
	if err != nil {
		return fmt.Errorf("delete employee fail err: %s", err)
	}
	r.tree = NewTree
	r.employees = TempEmployees
	r.targets = buildTargetsConfig(r.employees)

	return nil
}

//启动路由树
func (r *routerTree) Serve() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if r.serverState == ServerUp {
		return nil
	}

	r.server = &http.Server{}
	r.server.Handler = r

	ln, err := listener.ListenerTCP(r.listenPort, n)
	if err != nil {

	}
	go func(srv *http.Server) {
		err := srv.Serve(ln)
		log.Error(err)
	}(r.server)
	r.serverState = ServerUp

	return nil
}

func (r *routerTree) ShutDown() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if r.serverState == ServerDown {
		return nil
	}

	listener_manager.DeleteTCPListener(r.listenPort)
	r.server.Shutdown(context.Background())
	r.serverState = ServerDown
	return nil
}

func (r *routerTree) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//TODO

	target, has := r.tree.Match(request)
	if !has {
		return
	}
	result := map[string]interface{}{
		"location": r.targets[target].Location(),
		"host":     r.targets[target].Host(),
		"header":   r.targets[target].Header(),
		"query":    r.targets[target].Query(),
	}
	data, _ := json.Marshal(result)
	writer.Write(data)
}

func cloneEmployees(HES map[string]*httpEmployee) map[string]*httpEmployee {
	NewHES := make(map[string]*httpEmployee)
	for id, HE := range HES {
		NewHES[id] = HE
	}
	return NewHES
}
