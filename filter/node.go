package filter

import http_service "github.com/eolinker/eosc/http-service"

type _ChainNode struct {
	filter http_service.IFilter
	next   http_service.IChain
}

func createNode(filters []http_service.IFilter, end http_service.IChain) *_ChainNode {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return &_ChainNode{filter: filters[0], next: end}

	}
	return &_ChainNode{filter: filters[0], next: createNode(filters[1:], end)}
}
func (c *_ChainNode) DoChain(ctx http_service.IHttpContext) error {

	if c == nil {
		return nil
	}
	if c.filter == nil {
		return nil
	}
	return c.filter.DoFilter(ctx, c.next)
}
