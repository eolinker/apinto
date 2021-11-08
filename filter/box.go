package filter

import "github.com/eolinker/eosc/http"

type Box struct {
	filter http.IFilter
}

func NewBox(filter http.IFilter) *Box {
	return &Box{filter: filter}

}

func (b *Box) DoFilter(ctx http.IHttpContext, next http.IChain) (err error) {
	if b.filter != nil {
		return b.filter.DoFilter(ctx, next)
	}
	return next.DoFilter(ctx)
}

func (b *Box) reset(filter http.IFilter) {
	b.filter = filter
}
