package manager

import (
	"github.com/eolinker/apinto/router"
	grpc_router "github.com/eolinker/apinto/router/grpc-router"
)

var _ IRouterData = (*RouterData)(nil)

type IRouterData interface {
	Set(id string, port int, hosts []string, service string, method string, append []AppendRule, router router.IRouterHandler) IRouterData
	Delete(id string) IRouterData
	Parse() (router.IMatcher, error)
}
type RouterData struct {
	data map[string]*Router
}

func (rs *RouterData) Parse() (router.IMatcher, error) {
	root := grpc_router.NewRoot()
	for _, v := range rs.data {
		err := root.Add(v.Id, v.GrpcRouter, v.Port, v.Hosts, v.Service, v.Method, v.Appends)
		if err != nil {
			return nil, err
		}
	}
	return root.Build(), nil
}

func (rs *RouterData) set(r *Router) *RouterData {
	rs.data[r.Id] = r
	return rs
}
func (rs *RouterData) Set(id string, port int, hosts []string, service string, method string, append []AppendRule, router router.IRouterHandler) IRouterData {
	r := &Router{
		Id:         id,
		Port:       port,
		Hosts:      hosts,
		Method:     method,
		Service:    service,
		Appends:    append,
		GrpcRouter: router,
	}
	return rs.clone(1).set(r)
}

func (rs *RouterData) Delete(id string) IRouterData {

	return rs.clone(0).delete(id)
}
func (rs *RouterData) delete(id string) IRouterData {
	delete(rs.data, id)
	return rs
}
func (rs *RouterData) clone(delta int) *RouterData {
	if delta < 0 {
		delta = 0
	}
	if rs == nil || len(rs.data) == 0 {
		return &RouterData{data: make(map[string]*Router, 1)}
	}

	data := make(map[string]*Router, len(rs.data)+delta)
	for k, v := range rs.data {
		data[k] = v
	}
	return &RouterData{data: data}
}
