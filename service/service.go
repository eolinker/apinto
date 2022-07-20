package service

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/context"
	http_service "github.com/eolinker/eosc/context/http-context"
)

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/service.service.IService"
}

//IService github.com/eolinker/apinto/service.service.IService
type IService interface {
	context.IChain
}
type IServiceCreate interface {
	Create(id string, configs map[string]*plugin.Config) IService
}

//IRouterEndpoint 实现了返回路由规则信息方法的接口，如返回location、Host、Header、Query
type IRouterEndpoint interface {
	Location() (checker.Checker, bool)
	Header(name string) (checker.Checker, bool)
	Query(name string) (checker.Checker, bool)
	Headers() []string
	Queries() []string
}

type routerEndpointKey struct{}

var RouterEndpointKey = routerEndpointKey{}

////IServiceDetail 实现了返回服务信息方法的接口，如返回服务名，服务描述，重试次数间等..
//type IServiceDetail interface {
//	Name() string
//	Desc() string
//	Retry() int
//	Timeout() time.Duration
//	Scheme() string
//	ProxyAddr() string
//}

func EndpointFromContext(ctx http_service.IHttpContext) (IRouterEndpoint, bool) {
	value := ctx.Value(RouterEndpointKey)
	if value != nil {
		ep, ok := value.(IRouterEndpoint)
		return ep, ok
	}
	return nil, false
}
func AddEndpoint(ctx http_service.IHttpContext, endpoint IRouterEndpoint) {
	ctx.WithValue(RouterEndpointKey, endpoint)
}
