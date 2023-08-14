package dynamic_params

import (
	"github.com/eolinker/eosc/eocontext/http-context"
)

type IDynamicFactory interface {
	Create(name string, value []string) (IDynamicDriver, error)
}

type IDynamicDriver interface {
	Name() string
	Generate(ctx http_context.IHttpContext, contentType string, args ...interface{}) (interface{}, error)
}
