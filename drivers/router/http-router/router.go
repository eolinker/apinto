package http_router

import (
	router_http "github.com/eolinker/apinto/router/router-http"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

//Router http路由驱动实例结构体，实现了worker接口
type Router struct {
	id   string
	name string
	port int

	driver  *HTTPRouterDriver
	handler *RouterHandler
}

//Reset 重置http路由配置
func (r *Router) create(cf *DriverConfig, target service.IServiceCreate) (*RouterHandler, error) {

	newConf := getConfig(cf)
	newConf.ID = r.id
	newConf.Name = r.name
	if cf.Disable {
		return NewDisableHandler(newConf), nil
	}
	serviceHandler := target.Create(r.id, cf.Plugins)
	handler := NewRouterHandler(newConf, serviceHandler)
	return handler, nil
}
func (r *Router) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cf, ser, err := r.driver.check(conf, workers)
	if err != nil {
		return err
	}

	routerHandler, err := r.create(cf, ser)
	if err != nil {
		return err
	}

	err = router_http.Add(cf.Listen, r.id, routerHandler.routerConfig)
	if err != nil {
		routerHandler.Destroy()
		return err
	}

	if cf.Listen != r.port {
		router_http.Del(r.port, r.id)
	}

	r.port = cf.Listen
	r.handler = routerHandler
	return nil
}

//CheckSkill 技能检查
func (r *Router) CheckSkill(skill string) bool {
	return false
}

//Id 返回workerID
func (r *Router) Id() string {
	return r.id
}

//Start 启动路由worker，将路由实例加入到路由树中
func (r *Router) Start() error {
	log.Debug("router:start")
	return router_http.Add(r.port, r.id, r.handler.routerConfig)
}

//Stop 停止路由worker，将路由实例从路由树中删去
func (r *Router) Stop() error {
	return router_http.Del(r.port, r.id)
}

func getConfig(cf *DriverConfig) *router_http.Config {

	rules := make([]router_http.Rule, 0, len(cf.Rules))
	for _, r := range cf.Rules {
		rr := router_http.Rule{
			Location: r.Location,
			Header:   make([]router_http.HeaderItem, 0, len(r.Header)),
			Query:    make([]router_http.QueryItem, 0, len(r.Query)),
		}
		for k, v := range r.Header {
			rr.Header = append(rr.Header, router_http.HeaderItem{
				Name:    k,
				Pattern: v,
			})
		}
		for k, v := range r.Query {
			rr.Query = append(rr.Query, router_http.QueryItem{
				Name:    k,
				Pattern: v,
			})
		}
		rules = append(rules, rr)
	}
	// 配置里的Host或Method字段若为空，则默认该路由允许任何的host或method值
	hosts := cf.Host
	if len(hosts) == 0 {
		hosts = []string{"*"}
	}
	methods := cf.Method
	if len(methods) == 0 {
		methods = []string{"*"}
	}

	return &router_http.Config{
		//Cert:    certs,
		Methods: methods,
		Hosts:   hosts,
		Rules:   rules,
	}

}
