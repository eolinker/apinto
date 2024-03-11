package manager

import "github.com/eolinker/apinto/router"

type Router struct {
	Id          string
	Port        int
	Protocols   []string
	Hosts       []string
	Method      []string
	Path        string
	Appends     []AppendRule
	HttpHandler router.IRouterHandler
}
