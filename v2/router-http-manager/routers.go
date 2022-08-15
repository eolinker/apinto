package router_http_manager

import (
	"github.com/eolinker/apinto/v2/router"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type Routers struct {
}

func (r *Routers) Match(port int, request http_service.IRequestReader) (router.IRouterHandler, bool) {

	return nil, false
}
