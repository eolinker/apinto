package service

import (
	"net/http"
	"time"
)

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku-eosc/service.service.IService"
}

//IService github.com/eolinker/goku-eosc/service.service.IService
type IService interface {
	Handle(w http.ResponseWriter, r *http.Request, router IRouterRule) error
}

//IRouterRule 实现了返回路由规则信息方法的接口，如返回location、Host、Header、Query
type IRouterRule interface {
	Location() string
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
