package cache_strategy

import (
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	RegisterActuator(newActuator())
}

type actuatorHttp struct {
}

func newActuator() *actuatorHttp {
	return &actuatorHttp{}
}

func (hd *actuatorHttp) Assert(ctx eocontext.EoContext) bool {
	_, err := http_service.Assert(ctx)
	if err != nil {
		return false
	}
	return true
}

func (hd *actuatorHttp) Check(ctx eocontext.EoContext, handlers []*CacheValidTimeHandler) error {
	//httpCtx, err := http_service.Assert(ctx)
	//if err != nil {
	//	return err
	//}

	return nil
}

type Set map[string]struct{}

func newSet(l int) Set {
	return make(Set, l)
}
func (s Set) Has(key string) bool {
	_, has := s[key]
	return has
}
func (s Set) Add(key string) {
	s[key] = struct{}{}
}
