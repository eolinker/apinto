package grey_strategy

import "github.com/eolinker/eosc/eocontext"

type greyFlow struct {
	flowRobin *Robin
}

// Match 按流量计算
func (r *greyFlow) Match(ctx eocontext.EoContext) bool {
	flow := r.flowRobin.Select()
	return flow.GetId() == 1
}
