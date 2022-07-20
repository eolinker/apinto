package filter

import (
	"github.com/eolinker/eosc/context"
)

type IChainReset interface {
	Reset(filters ...context.IFilter)
}

type IChain interface {
	context.IChain
	ToFilter() context.IFilter
	Append(filters ...context.IFilter) IChain
	Insert(filters ...context.IFilter) IChain
}

type IChainHandler interface {
	IChain
	IChainReset
}

func NewChain(filters []context.IFilter) IChainHandler {
	return createHandler(filters)
}
