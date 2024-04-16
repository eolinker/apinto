package access_relational

import (
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"strconv"
	"time"
)

func (a *AccessRelational) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {

	if len(a.rules) == 0 {
		return next.DoChain(ctx)
	}
	now := time.Now().UnixMilli()
	for _, rule := range a.rules {
		key := rule.key.Metrics(ctx)
		field := rule.field.Metrics(ctx)
		v, has := a.data.Get(key, field)
		if !has {
			// 规则不存在
			continue
		}
		timestamp, _ := strconv.ParseInt(v, 10, 64)
		if timestamp <= 0 || timestamp > now {
			// 校验通过
			return next.DoChain(ctx)
		}
	}

	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	a.response.Response(httpContext)
	return nil
}

func (a *AccessRelational) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(a, ctx, next)
}
