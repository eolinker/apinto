package filter

import (
	"github.com/eolinker/eosc/context"
)

var _ context.IFilter = (*_ChainFilter)(nil)

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

func ToFilter(filters []context.IFilter) *_ChainFilter {
	c := &_ChainFilter{}
	c.Reset(filters...)
	return c
}

func (c *_ChainFilter) Reset(filters ...context.IFilter) {
	c.startNode = createNode(filters, c)
}

func (c *_ChainFilter) DoChain(ctx context.Context) error {
	value := ctx.Value(c)
	if value == nil {
		return nil
	}
	if next, ok := value.(context.IChain); ok {
		return next.DoChain(ctx)
	}
	return nil
}

func (c *_ChainFilter) DoFilter(ctx context.Context, next context.IChain) (err error) {

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
