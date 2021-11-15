package filter

import http_service "github.com/eolinker/eosc/http-service"

var _ IChainHandler = (*_ChainHandler)(nil)

type _ChainHandler struct {
	orgFilter    *_ChainFilter
	resetHandler IChainReset
}

func (c *_ChainHandler) Destroy() {
	c.orgFilter.Destroy()
}

func createHandler(filters []http_service.IFilter) *_ChainHandler {
	orgFilter := ToFilter(filters)
	return &_ChainHandler{
		orgFilter:    orgFilter,
		resetHandler: orgFilter,
	}

}

func (c *_ChainHandler) ToFilter() http_service.IFilter {
	return c.orgFilter
}

func (c *_ChainHandler) DoChain(ctx http_service.IHttpContext) error {
	return c.orgFilter.DoFilter(ctx, nil)
}

func (c *_ChainHandler) Append(filters ...http_service.IFilter) IChain {
	pre := c.ToFilter()
	fs := make([]http_service.IFilter, 0, len(filters)+1)
	fs = append(fs, pre)
	fs = append(fs, filters...)
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Insert(filters ...http_service.IFilter) IChain {
	pre := c.ToFilter()
	fs := make([]http_service.IFilter, 0, len(filters)+1)
	fs = append(fs, filters...)
	fs = append(fs, pre)
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Reset(filters ...http_service.IFilter) {
	if c.resetHandler == nil {
		c.orgFilter.Reset(filters...)
		return
	}
	c.resetHandler.Reset(filters...)
}
