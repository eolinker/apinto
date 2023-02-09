package manager

import "github.com/eolinker/apinto/router"

type Router struct {
	Id      string
	Port    int
	Hosts   []string
	Service string
	Method  string
	Appends []AppendRule
	Router  router.IRouterHandler
}
