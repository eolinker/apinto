package http_router

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	router_http "github.com/eolinker/goku/router/router-http"
	"github.com/eolinker/goku/service"
)

//Router http路由驱动实例结构体，实现了worker接口
type Router struct {
	id   string
	name string
	port int
	conf *router_http.Config

	driver *HTTPRouterDriver
}

//func (r *Router) Ports() []int {
//
//	return []int{r.port}
//}

//Reset 重置http路由配置
func (r *Router) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cf, target, err := r.driver.check(conf, workers)
	if err != nil {
		return err
	}

	newConf := getConfig(target, cf)
	newConf.ID = r.id
	newConf.Name = r.name
	err = router_http.Add(cf.Listen, r.id, newConf)
	if err != nil {
		return err
	}

	if cf.Listen != r.port {
		router_http.Del(r.port, r.id)
	}

	r.port = cf.Listen
	r.conf = newConf

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
	return router_http.Add(r.port, r.id, r.conf)
}

//Stop 停止路由worker，将路由实例从路由树中删去
func (r *Router) Stop() error {
	return router_http.Del(r.port, r.id)
}

func getConfig(target service.IService, cf *DriverConfig) *router_http.Config {

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

	//protocol := "http"
	//if cf.Protocol == "https" {
	//	protocol = "https"
	//}

	//certs := make([]router_http.Cert, 0, len(cf.Cert))
	//for _, c := range cf.Cert {
	//	certs = append(certs, router_http.Cert{Key: c.Key, Crt: c.Crt})
	//}

	return &router_http.Config{
		//Cert:    certs,
		Methods: methods,
		Hosts:   hosts,
		Target:  target,
		Rules:   rules,
	}

}

//NewRouter 创建http路由驱动实例
func NewRouter(id, name string, c *DriverConfig, target service.IService, driver *HTTPRouterDriver) *Router {
	conf := getConfig(target, c)
	conf.ID = id
	conf.Name = name

	return &Router{
		id:     id,
		name:   name,
		port:   c.Listen,
		conf:   conf,
		driver: driver,
	}
}
