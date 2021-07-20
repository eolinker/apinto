package router_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	listener_manager "github.com/eolinker/goku-eosc/listener-manager"
)

type RouterManagerDemo struct {
	servers map[int]*RouterServerDemo
	locker  sync.RWMutex
}

func NewRouterManagerDemo() *RouterManagerDemo {
	return &RouterManagerDemo{locker: sync.RWMutex{}, servers: make(map[int]*RouterServerDemo)}
}

func (r *RouterManagerDemo) Get(port int) (*RouterServerDemo, error) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	if v, ok := r.servers[port]; ok {
		return v, nil
	}
	return nil, errors.New("")
}

func (r *RouterManagerDemo) Set(port int, server *RouterServerDemo) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.servers[port] = server
}

type RouterServerDemo struct {
	port   int
	server *http.Server
	hosts  []string
}

func (r *RouterServerDemo) AppendHost(host ...string) {
	r.hosts = append(r.hosts, host...)
}

func (r *RouterServerDemo) Shutdown(ctx context.Context) error {
	r.server.Shutdown(ctx)
	return nil
}
func NewRouterServerDemo(port int, hosts []string) (*RouterServerDemo, error) {
	ln, err := listener_manager.GetTCPListener(port)
	if err == nil {
		return nil, errors.New("the port is already occupied")
	}
	ln, err = listener_manager.NewTCPListener("0.0.0.0", port)
	if err != nil {
		return nil, err
	}
	listener_manager.SetTCPListener(port, ln)
	r := &RouterServerDemo{port: port, hosts: hosts, server: &http.Server{}}
	r.server.Handler = r
	go r.server.Serve(ln)

	return r, nil
}

func (r *RouterServerDemo) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	result := map[string]interface{}{
		"port": r.port,
		"host": r.hosts,
	}
	b, _ := json.Marshal(result)
	writer.Write(b)
}

func TestRouter(t *testing.T) {
	rm := NewRouterManagerDemo()
	r, err := rm.Get(8080)
	if err != nil {
		r, err = NewRouterServerDemo(8080, []string{"www.baidu.com", "www.eolinker.com"})
		if err != nil {
			return
		}
	}

	rm.Set(8080, r)
	newR, err := rm.Get(8080)
	if err == nil {
		newR.AppendHost("www.goku.com", "www.apibee.com")
	}
	select {}
}
