package manager

import "github.com/eolinker/apinto/router"

type Router struct {
	Id          string
	Port        int
	Hosts       []string
	Method      []string
	Path        string
	Appends     []AppendRule
	HttpHandler router.IRouterHandler
}
