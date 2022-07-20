package filter

import (
	"github.com/eolinker/eosc/context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
)

var _ IChainHandler = (*_ChainHandler)(nil)

type _ChainHandler struct {
	orgFilter    *_ChainFilter
	resetHandler IChainReset
}

func (c *_ChainHandler) Destroy() {
	orgFilter := c.orgFilter
	if orgFilter != nil {
		c.orgFilter = nil
		orgFilter.Destroy()
	}

}

func createHandler(filters []context.IFilter) *_ChainHandler {
	orgFilter := ToFilter(filters)
	return &_ChainHandler{
		orgFilter:    orgFilter,
		resetHandler: orgFilter,
	}

}

func (c *_ChainHandler) ToFilter() context.IFilter {
	return c.orgFilter
}

func (c *_ChainHandler) DoChain(ctx context.Context) error {
	log.Debug("do chain handler: ", c, config.TypeNameOf(c.orgFilter))
	orgFilter := c.orgFilter
	if orgFilter != nil {
		return orgFilter.DoFilter(ctx, nil)
	}
	return nil
}

func (c *_ChainHandler) Append(filters ...context.IFilter) IChain {
	pre := c.ToFilter()
	fs := make([]context.IFilter, 0, len(filters)+1)
	if pre != nil {
		fs = append(fs, pre)
	}
	fs = append(fs, filters...)
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Insert(filters ...context.IFilter) IChain {
	pre := c.ToFilter()

	fs := make([]context.IFilter, 0, len(filters)+1)
	fs = append(fs, filters...)
	if pre != nil {
		fs = append(fs, pre)
	}
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Reset(filters ...context.IFilter) {

	if c.resetHandler != nil {

		c.resetHandler.Reset(filters...)
		return
	}
	filter := c.orgFilter
	if filter != nil {
		filter.Reset(filters...)
	} else {
		c.orgFilter = ToFilter(filters)
	}
	return
}
