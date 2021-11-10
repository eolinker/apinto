package filter

import http_service "github.com/eolinker/eosc/http-service"

type IChainReset interface {
	Reset(filters ...http_service.IFilter)
}

type IChain interface {
	http_service.IChain
	ToFilter() http_service.IFilter
	Append(filters ...http_service.IFilter) IChain
	Insert(filters ...http_service.IFilter) IChain
}

type IChainHandler interface {
	IChain
	IChainReset
}

func NewChain(filters []http_service.IFilter) IChainHandler {
	return createHandler(filters)
}
