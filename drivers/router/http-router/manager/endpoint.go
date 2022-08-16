package manager

import http_router "github.com/eolinker/apinto/router/http-router"

type Router struct {
	Id          string
	Port        int
	Hosts       []string
	Method      []string
	Path        string
	Appends     []AppendRule
	HttpHandler http_router.IRouterHandler
}
