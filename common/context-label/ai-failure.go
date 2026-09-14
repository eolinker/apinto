package context_label

import "github.com/eolinker/eosc/eocontext"

const (
	LabelAIFailure = "ai_failure"
)

// SetAIFailure 由 AI 供应商在判定调用失败时写入该标签
func SetAIFailure(ctx eocontext.EoContext, failed bool) {
	if failed {
		ctx.SetLabel(LabelAIFailure, "true")
	} else {
		ctx.SetLabel(LabelAIFailure, "false")
	}
}

// IsAIFailure 获取当前请求是否由 AI 供应商标记为失败
func IsAIFailure(ctx eocontext.EoContext) bool {
	return ctx.GetLabel(LabelAIFailure) == "true"
}

// HasAIFailureLabel 判断上下文是否显式设置了 AI 失败标签（不为空）
func HasAIFailureLabel(ctx eocontext.EoContext) bool {
	return ctx.GetLabel(LabelAIFailure) != ""
}
