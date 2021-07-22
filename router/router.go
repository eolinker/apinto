package router

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/eolinker/eosc"
)

type IRouterHttpFactory interface {
}

type IRouterManager interface {
	Delete(port int, id string) error
	StartAllServer()
	ShutDownAllServer()
	StartServer(port int) error
	ShutDownServer(port int) error
}

type IRouterHandler interface {
	Match(request *http.Request) (string, bool)
}

type IRouter interface {
	Delete(id string) error
	Serve() error
	ShutDown() error
}

type IRouterRule interface {
	Location() string
	Host() string
	Header() map[string]string
	Query() url.Values
}

func NewRouter(name string) *Router {
	return &Router{
		id:   fmt.Sprintf("%s:%s_%s:%s", group, profession, name, version),
		name: name,
	}
}

//Router 路由模块
type Router struct {
	id    string
	name  string
	label string
}

func (r *Router) ConfigType() reflect.Type {
	panic("implement me")
}

func (r *Router) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	panic("implement me")
}

func (r *Router) Name() string {
	return r.name
}

func (r *Router) Check(config string) error {
	return nil
	panic("implement me")
}

func (r *Router) Render() eosc.Render {
	panic("implement me")
}
