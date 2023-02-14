package manager

import "github.com/eolinker/apinto/router"

type Router struct {
	Id         string
	Port       int
	Service    string
	Method     string
	Appends    []AppendRule
	GrpcRouter router.IRouterHandler
}
