package filter

import (
	"github.com/eolinker/eosc/eocontext"
)

type IChainReset interface {
	Reset(filters ...eocontext.IFilter)
}

type IChainHandler interface {
	IChainReset
}

func NewChain(filters []eocontext.IFilter) IChainHandler {
	return createHandler(filters)
}
