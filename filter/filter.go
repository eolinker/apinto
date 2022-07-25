package filter

import "github.com/eolinker/eosc/eocontext"

var _ eocontext.IFilter = (*_ChainFilter)(nil)

type _ChainFilter struct {
	startNode *_ChainNode
}

func (c *_ChainFilter) Destroy() {
	startNode := c.startNode
	if startNode != nil {
		c.startNode = nil
		startNode.Destroy()
	}

}

func ToFilter(filters []eocontext.IFilter) *_ChainFilter {
	c := &_ChainFilter{}
	c.Reset(filters...)
	return c
}

func (c *_ChainFilter) Reset(filters ...eocontext.IFilter) {
	c.startNode = createNode(filters, c)
}

func (c *_ChainFilter) DoChain(ctx eocontext.EoContext) error {
	value := ctx.Value(c)
	if value == nil {
		return nil
	}
	if next, ok := value.(eocontext.IChain); ok {
		return next.DoChain(ctx)
	}
	return nil
}

func (c *_ChainFilter) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {

	startNode := c.startNode
	if startNode != nil {
		if next != nil {
			ctx.WithValue(c, next)
		}
		return startNode.DoChain(ctx)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}
