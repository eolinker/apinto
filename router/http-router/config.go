package http_router

import "net/http"

type Config struct {
	name    string
	port    int
	Rules   []RouterRule
	host    []string
	service http.Handler
}

type RouterRule struct {
	location string
	header   map[string]string
	query    map[string]string
}

type RouterWork struct {
	Service http.Handler
	Config  Config
}

//func (r *RouterWork) Start() error {
//	routerManager.Add(r.Config, r.Service)
//}
//
//func (r *RouterWork) Stop() error {
//	routerManager.Del(r.Config, r.Service)
//}
