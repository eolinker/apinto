// github.com/eolinker/goku-eosc/service.service.IService

package service

import (
	"net/http"
	"net/url"
	"time"
)

func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku-eosc/service.service.IService"
}

// IService github.com/eolinker/goku-eosc/service.service.IService
type IService interface {
	Handle(w http.ResponseWriter, r *http.Request, router IRouterRule) error
}

type IRouterRule interface {
	Location() string
	Host() string
	Header() map[string]string
	Query() url.Values
}

type IServiceDetail interface {
	Name() string
	Desc() string
	Retry() int
	Timeout() time.Duration
	Scheme() string
	ProxyAddr() string
}
