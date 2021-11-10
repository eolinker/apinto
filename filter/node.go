package filter

import "github.com/eolinker/eosc/http"

type _ChainNode struct {
	filter http.IFilter
	next   http.IChain
}

func createNode(filters []http.IFilter, end http.IChain) *_ChainNode {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return &_ChainNode{filter: filters[0], next: end}

	}
	return &_ChainNode{filter: filters[0], next: createNode(filters[1:], end)}
}
func (c *_ChainNode) DoChain(ctx http.IHttpContext) error {

	if c == nil {
		return nil
	}
	if c.filter == nil {
		return nil
	}
	return c.filter.DoFilter(ctx, c.next)
}
