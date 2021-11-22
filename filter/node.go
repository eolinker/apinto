package filter

import (
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
)

var _ http_service.IChain = (*_ChainNode)(nil)

type _ChainNode struct {
	filter http_service.IFilter
	next   http_service.IChain
}

func (c *_ChainNode) Destroy() {
	if c == nil {
		return
	}
	if c.filter != nil {
		c.filter.Destroy()
		c.filter = nil
	}
	if c.next != nil {
		c.next.Destroy()
		c.next = nil
	}
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
	log.Debug(" chain: ", c, "filter: ", eosc.TypeNameOf(c.filter))
	if c == nil {
		return nil
	}
	if c.filter == nil {
		return nil
	}
	return c.filter.DoFilter(ctx, c.next)
}
