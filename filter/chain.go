package filter

import "github.com/eolinker/eosc/http"

type ChainItem struct {
	filter http.IFilter
	next   http.IChain
}

func create(filters []http.IFilter) *ChainItem {
	if len(filters) == 0 {
		return &ChainItem{
			filter: nil,
			next:   nil,
		}
	}
	return &ChainItem{
		filter: filters[0],
		next:   create(filters[1:]),
	}
}
func (c *ChainItem) DoFilter(ctx http.IHttpContext) error {
	if c == nil {
		return nil
	}
	if c.filter != nil {
		err := c.filter.DoFilter(ctx, c.next)
		return err
	}

	return nil
}
