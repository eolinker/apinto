package http_router

import (
	"github.com/eolinker/eosc"
	router_http "github.com/eolinker/goku-eosc/router/router-http"
	"github.com/eolinker/goku-eosc/service"
)

type Router struct {
	id   string
	name string
	port int
	conf *router_http.Config

	driver *HttpRouterDriver
}

func (r *Router) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cf, target, err := r.driver.check(conf, workers)
	if err != nil {
		return err
	}

	newConf := getConfig(target, cf)
	newConf.Id = r.id
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

func (r *Router) CheckSkill(skill string) bool {
	return false
}

func (r *Router) Id() string {
	return r.id
}

func (r *Router) Start() error {
	return router_http.Add(r.port, r.id, r.conf)
}

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
	hosts := cf.Host
	if len(hosts) == 0 {
		hosts = []string{"*"}
	}
	methods := cf.Method
	if len(methods) == 0 {
		methods = []string{"*"}
	}
	return &router_http.Config{
		//Id:     cf.ID,
		//Name:   cf.Name,
		Methods: methods,
		Hosts:   hosts,
		Target:  target,
		Rules:   rules,
	}

}
func NewRouter(id, name string, c *DriverConfig, target service.IService) *Router {
	conf := getConfig(target, c)
	conf.Id = id
	conf.Name = name

	return &Router{
		id:   id,
		name: name,
		port: c.Listen,
		conf: conf,
	}
}
