package filter

import "github.com/eolinker/eosc/http"

type _ChainHandler struct {
	orgFilter    *_ChainFilter
	resetHandler IChainReset
}

func createHandler(filters []http.IFilter) *_ChainHandler {
	orgFilter := toFilter(filters)
	return &_ChainHandler{
		orgFilter:    orgFilter,
		resetHandler: orgFilter,
	}

}

func (c *_ChainHandler) ToFilter() http.IFilter {
	return c.orgFilter
}

func (c *_ChainHandler) DoChain(ctx http.IHttpContext) error {
	return c.orgFilter.DoFilter(ctx, nil)
}

func (c *_ChainHandler) Append(filters ...http.IFilter) IChain {
	pre := c.ToFilter()
	fs := make([]http.IFilter, 0, len(filters)+1)
	fs = append(fs, pre)
	fs = append(fs, filters...)
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Insert(filters ...http.IFilter) IChain {
	pre := c.ToFilter()
	fs := make([]http.IFilter, 0, len(filters)+1)
	fs = append(fs, filters...)
	fs = append(fs, pre)
	n := createHandler(fs)
	n.resetHandler = c.resetHandler
	return n
}

func (c *_ChainHandler) Reset(filters ...http.IFilter) {
	if c.resetHandler == nil {
		c.orgFilter.Reset(filters...)
		return
	}
	c.resetHandler.Reset(filters...)
}
