package service

import (
	"time"

	http_service "github.com/eolinker/eosc/http-service"

	http_context "github.com/eolinker/goku/node/http-context"
)

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku/service.service.IService"
}

//IService github.com/eolinker/goku/service.service.IService
type IService interface {
	Handle(ctx *http_context.Context, router IRouterEndpoint) error
}

//IRouterEndpoint 实现了返回路由规则信息方法的接口，如返回location、Host、Header、Query
type IRouterEndpoint interface {
	Location() (http_service.Checker, bool)
	Header(name string) (http_service.Checker, bool)
	Query(name string) (http_service.Checker, bool)
	Headers() []string
	Queries() []string
}

//IServiceDetail 实现了返回服务信息方法的接口，如返回服务名，服务描述，重试次数间等..
type IServiceDetail interface {
	Name() string
	Desc() string
	Retry() int
	Timeout() time.Duration
	Scheme() string
	ProxyAddr() string
}
