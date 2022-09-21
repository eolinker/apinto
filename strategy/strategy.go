package strategy

import eoscContext "github.com/eolinker/eosc/eocontext"

type IStrategyHandler interface {
	Strategy(ctx eoscContext.EoContext, next eoscContext.IChain) error
}

type IFilter interface {
	Check(ctx eoscContext.EoContext) bool
}

type IFilters []IFilter

func (fs IFilters) Check(ctx eoscContext.EoContext) bool {
	for _, f := range fs {
		if !f.Check(ctx) {
			return false
		}
	}
	return true
}
