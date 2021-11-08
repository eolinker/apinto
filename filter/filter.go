package filter

import (
	"github.com/eolinker/eosc/http"
)

type IChain interface {
	http.IFilterChain
	Append(filters ...http.IFilter) IChain
	Insert(filters ...http.IFilter) IChain
	Merge(chain IChain) IChain
}

type Chain struct {
	*ChainItem
	filters []http.IFilter
}

func Create(filters []http.IFilter) IChain {
	return &Chain{
		ChainItem: create(filters),
		filters:   filters,
	}
}

func (c *Chain) Append(filters ...http.IFilter) IChain {
	nf := make([]http.IFilter, 0, len(filters)+len(c.filters))
	nf = append(nf, c.filters...)
	nf = append(nf, filters...)
	return Create(nf)
}

func (c *Chain) Insert(filters ...http.IFilter) IChain {
	nf := make([]http.IFilter, 0, len(filters)+len(c.filters))
	nf = append(nf, filters...)
	nf = append(nf, c.filters...)
	return Create(nf)
}

func (c *Chain) Merge(chain IChain) IChain {

	return c.Append(&NextFilter{
		next: chain,
	})
}
