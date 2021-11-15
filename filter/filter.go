package filter

import http_service "github.com/eolinker/eosc/http-service"

type _ChainFilter struct {
	startNode *_ChainNode
}

func toFilter(filters []http_service.IFilter) *_ChainFilter {
	c := &_ChainFilter{}
	c.Reset(filters...)
	return c
}

func (c *_ChainFilter) Reset(filters ...http_service.IFilter) {

	c.startNode = createNode(filters, c)
}

func (c *_ChainFilter) DoChain(ctx http_service.IHttpContext) error {
	value := ctx.Value(c)
	if value == nil {
		return nil
	}
	if next, ok := value.(http_service.IChain); ok {

		return next.DoChain(ctx)
	}
	return nil
}

func (c *_ChainFilter) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {

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
