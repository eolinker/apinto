package filter

import (
	"github.com/eolinker/eosc/http"
)

type ChainNode struct {
	filter http.IFilter
	next   http.IChain
}

func (c *ChainNode) DoChain(ctx http.IHttpContext) error {

	if c == nil {
		return nil
	}
	if c.filter == nil {
		return nil
	}
	return c.filter.DoFilter(ctx, c.next)
}

func createNode(filters []http.IFilter, end http.IChain) *ChainNode {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return &ChainNode{filter: filters[0], next: end}

	}
	return &ChainNode{filter: filters[0], next: createNode(filters[1:], end)}
}

type ChainFilter struct {
	startNode *ChainNode
}

func NewChainFilter(filters []http.IFilter) *ChainFilter {
	c := &ChainFilter{}
	c.Reset(filters...)
	return c
}

func (c *ChainFilter) Reset(filters ...http.IFilter) {

	c.startNode = createNode(filters, c)
}

func (c *ChainFilter) DoChain(ctx http.IHttpContext) error {
	value := ctx.Value(c)
	if value == nil {
		return nil
	}
	if next, ok := value.(http.IChain); ok {

		return next.DoChain(ctx)
	}
	return nil
}

func (c *ChainFilter) DoFilter(ctx http.IHttpContext, next http.IChain) (err error) {

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

type ChainHandler struct {
	orgFilter    *ChainFilter
	resetHandler IChainReset
}

func (c *ChainHandler) ToFilter() http.IFilter {
	return c.orgFilter
}

func (c *ChainHandler) DoChain(ctx http.IHttpContext) error {
	return c.orgFilter.DoFilter(ctx, nil)
}

func (c *ChainHandler) Append(filters ...http.IFilter) IChain {
	pre := c.orgFilter
	fs := make([]http.IFilter, 0, len(filters)+1)
	fs = append(fs, pre)
	fs = append(fs, filters...)
	n := newChainHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *ChainHandler) Insert(filters ...http.IFilter) IChain {
	pre := c.orgFilter
	fs := make([]http.IFilter, 0, len(filters)+1)
	fs = append(fs, filters...)
	fs = append(fs, pre)
	n := newChainHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *ChainHandler) Reset(filters ...http.IFilter) {
	if c.resetHandler == nil {
		c.orgFilter.Reset(filters...)
		return
	}
	c.resetHandler.Reset(filters...)
}

func NewChainHandler(filters []http.IFilter) IChainHandler {
	return newChainHandler(filters)
}
func newChainHandler(filters []http.IFilter) *ChainHandler {
	orgFilter := NewChainFilter(filters)
	return &ChainHandler{
		orgFilter:    orgFilter,
		resetHandler: orgFilter,
	}

}
