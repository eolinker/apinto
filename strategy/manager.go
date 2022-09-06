package strategy

type IStrategyManager interface {
	Set(name string, filter IFilter)
}
