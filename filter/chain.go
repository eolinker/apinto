package filter

import (
	"github.com/eolinker/eosc/eocontext"
)

type IChainReset interface {
	Reset(filters ...eocontext.IFilter)
}

type IChain interface {
	eocontext.IChain
	ToFilter() eocontext.IFilter
	Append(filters ...eocontext.IFilter) IChain
	Insert(filters ...eocontext.IFilter) IChain
}

type IChainHandler interface {
	IChain
	IChainReset
}

func NewChain(filters []eocontext.IFilter) IChainHandler {
	return createHandler(filters)
}
