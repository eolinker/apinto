package filter

import (
	"github.com/eolinker/eosc/eocontext"
)

var _ IChainHandler = (*_ChainHandler)(nil)

type _ChainHandler struct {
	filters eocontext.Filters
}

func (c *_ChainHandler) Destroy() {
	c.filters.Destroy()
}

func createHandler(filters []eocontext.IFilter) *_ChainHandler {

	return &_ChainHandler{
		filters: filters,
	}

}

func (c *_ChainHandler) DoChain(ctx eocontext.EoContext) error {

	orgFilter := c.filters

	return orgFilter.DoChain(ctx)

}

func (c *_ChainHandler) Reset(filters ...eocontext.IFilter) {
	c.filters = filters
}
