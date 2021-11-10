package filter

import (
	"github.com/eolinker/eosc/http"
)

type IChainReset interface {
	Reset(filters ...http.IFilter)
}

type IChain interface {
	http.IChain
	Append(filters ...http.IFilter) IChain
	Insert(filters ...http.IFilter) IChain
}
type IChainHandler interface {
	IChain
	IChainReset
	ToFilter() http.IFilter
}
