package quota_limiting_strategy

import eoscContext "github.com/eolinker/eosc/eocontext"

var (
	LabelUserOfResourceGroup = "user_of_resource_group"
)

func GetUserOfResourceGroup(ctx eoscContext.EoContext) string {
	return ctx.GetLabel(LabelUserOfResourceGroup)
}

func SetUserOfResourceGroup(ctx eoscContext.EoContext, value string) {
	ctx.SetLabel(LabelUserOfResourceGroup, value)
}
