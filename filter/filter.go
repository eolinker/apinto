package filter

import (
	"github.com/eolinker/eosc/http"
)

type _ChainFilter struct {
	startNode *_ChainNode
}

func toFilter(filters []http.IFilter) *_ChainFilter {
	c := &_ChainFilter{}
	c.Reset(filters...)
	return c
}

func (c *_ChainFilter) Reset(filters ...http.IFilter) {

	c.startNode = createNode(filters, c)
}

func (c *_ChainFilter) DoChain(ctx http.IHttpContext) error {
	value := ctx.Value(c)
	if value == nil {
		return nil
	}
	if next, ok := value.(http.IChain); ok {

		return next.DoChain(ctx)
	}
	return nil
}

func (c *_ChainFilter) DoFilter(ctx http.IHttpContext, next http.IChain) (err error) {

	if c.startNode != nil {
		if next != nil {
			ctx.WithValue(c, next)
		}
		return c.startNode.DoChain(ctx)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}
