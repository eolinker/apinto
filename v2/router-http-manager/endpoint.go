package router_http_manager

import "github.com/eolinker/apinto/v2/router"

type Router struct {
	Id          string
	Port        int
	Hosts       []string
	Method      []string
	Path        string
	Appends     []AppendRule
	HttpHandler router.IRouterHandler
}
