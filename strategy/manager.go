package strategy

import eoscContext "github.com/eolinker/eosc/eocontext"

type IStrategyManager interface {
	AddStrategyHandler(handler IStrategyHandler)
}

func AddStrategyHandler(handler IStrategyHandler) {

}

func Strategy(ctx eoscContext.EoContext, next eoscContext.IChain) error {

	return nil
}
