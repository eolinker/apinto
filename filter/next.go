package filter

import "github.com/eolinker/eosc/http"

type NextFilter struct {
	next http.IChain
}

func (n *NextFilter) DoFilter(ctx http.IHttpContext, next http.IFilterChain) error {
	if n.next == nil {
		return next.DoFilter(ctx)
	}
	if err := n.next.DoFilter(ctx); err != nil {
		return err
	}
	return next.DoFilter(ctx)
}
