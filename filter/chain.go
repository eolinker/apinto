package filter

import (
	"github.com/eolinker/eosc/http"
)

type IChainReset interface {
	Reset(filters ...http.IFilter)
}

type IChain interface {
	http.IChain
	ToFilter() http.IFilter
	Append(filters ...http.IFilter) IChain
	Insert(filters ...http.IFilter) IChain
}
type IChainHandler interface {
	IChain
	IChainReset
}

func NewChain(filters []http.IFilter) IChainHandler {
	return createHandler(filters)
}
