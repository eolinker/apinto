package access_relational

import (
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

func (w *AccessRelational) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {

	if len(w.rules) == 0 {
		return next.DoChain(ctx)
	}
	entry := http_entry.NewEntry(ctx)
	for _, rule := range w.rules {
		if !rule.Check(entry) {
			continue
		}
		return next.DoChain(ctx)

	}

	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	w.response.Response(httpContext)
	return nil
}

func (w *AccessRelational) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(w, ctx, next)
}
