package filter

import (
	"github.com/eolinker/eosc/http"
)

type Chain struct {
	filter http.IFilter
	next   http.IChain
}

func CreateChain(filters []http.IFilter) *Chain {
	if len(filters) > 0 {
		return NewChain(filters[0], CreateChain(filters[1:]))
	}
	return nil
}

func NewChain(filter http.IFilter, next http.IChain) *Chain {
	return &Chain{filter: filter, next: next}
}

func (c *Chain) DoFilter(ctx http.IHttpContext, endpoint http.IEndpoint) error {
	if c.filter != nil {
		err := c.filter.DoFilter(ctx, endpoint, c.next)
		return err
	}

	return nil
}

func (c *Chain) Append(filter http.IFilter) {
	if c.next == nil {
		c.next = NewChain(filter, nil)
	} else {
		c.next.Append(filter)
	}
}

func (c *Chain) Insert(filter http.IFilter) {
	next := NewChain(c.filter, c.next)
	c.filter = filter
	c.next = next
}
