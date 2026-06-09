package utils

import "github.com/eolinker/eosc/eocontext"

type BillingMode string

func (b BillingMode) String() string {
	return string(b)
}

const (
	LabelBillingMode      = "billing_mode"
	BillingModeImmediate  = "immediate"
	BillingModeTaskCreate = "task_create"
	BillingModeTaskQuery  = "task_query"
)

func GetBillingMode(ctx eocontext.EoContext) BillingMode {
	value := ctx.GetLabel(LabelBillingMode)

	return BillingMode(value)
}

func SetBillingMode(ctx eocontext.EoContext, mode BillingMode) {
	ctx.SetLabel(LabelBillingMode, mode.String())
}

func IsBillingMode(ctx eocontext.EoContext, mode BillingMode) bool {
	value := GetBillingMode(ctx)
	return value == mode
}
