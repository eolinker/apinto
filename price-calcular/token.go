package price_calcular

import (
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc/eocontext"
)

func GetPreTotalToken(ctx eocontext.EoContext, rules []*ProcessedRule) int64 {
	if len(rules) < 1 {
		return 0
	}
	if rules[0].billingMode != BillingModeToken {
		return 0
	}
	
	var inputToken, outputToken int64
	inputToken = context_label.GetPreInputToken(ctx)
	if inputToken == 0 {
		inputToken = PreDeductInputTokensText
	}
	
	// 预估输出 token 数：根据上下文 label "model_type" 分配
	modelType := ctxLabel(ctx, LabelModelType)
	
	switch modelType {
	case ModelTypeImage, ModelTypeVideo:
		outputToken = int64(PreDeductTokensImageVideo)
	default:
		// text 或未指定：按 10K token 计
		outputToken = int64(PreDeductOutputTokensText)
	}
	
	return inputToken + outputToken
}
