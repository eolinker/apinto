package filter

import (
	"github.com/eolinker/eosc/context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
)

var _ context.IChain = (*_ChainNode)(nil)

type _ChainNode struct {
	filter context.IFilter
	next   context.IChain
}

func (c *_ChainNode) Destroy() {
	if c == nil {
		return
	}
	filter := c.filter
	if filter != nil {
		c.filter = nil
		filter.Destroy()
	}
	next := c.next
	if next != nil {
		c.next = nil
		next.Destroy()
	}
}

func createNode(filters []context.IFilter, end context.IChain) *_ChainNode {

	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return &_ChainNode{filter: filters[0], next: end}

	}
	return &_ChainNode{filter: filters[0], next: createNode(filters[1:], end)}
}
func (c *_ChainNode) DoChain(ctx context.Context) error {
	log.Debug(" chain: ", c, "filter: ", config.TypeNameOf(c.filter))
	if c == nil {
		return nil
	}
	filter := c.filter
	if filter == nil {
		return nil
	}
	return filter.DoFilter(ctx, c.next)
}
