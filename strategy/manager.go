package strategy

import "github.com/eolinker/eosc/eocontext"

type IStrategyManager interface {
	AddStrategyHandler(handler eocontext.IFilter)
	Strategy(ctx eocontext.EoContext, next eocontext.IChain) error
}

var (
	handlers eocontext.Filters
)

func AddStrategyHandler(handler eocontext.IFilter) {
	handlers = append(handlers, handler)
}

func Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return eocontext.DoChain(ctx, handlers, eocontext.ToFilter(next))
}
