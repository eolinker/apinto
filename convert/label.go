package convert

import "github.com/eolinker/eosc/eocontext"

var (
	AIModelInputTokenLabel  = "ai_model_input_token"
	AIModelOutputTokenLabel = "ai_model_output_token"
	AIModelTotalTokenLabel  = "ai_model_total_token"
	AIModelModeLabel        = "ai_model_mode"
	AIModelLabel            = "ai_model"
	AIProviderLabel         = "ai_provider"
	AIModelStatusLabel      = "ai_model_status"
)

func valueInt(ctx eocontext.EoContext, label string) int {
	value := ctx.Value(label)
	if v, ok := value.(int); ok {
		return v
	}
	return 0
}

func valueString(ctx eocontext.EoContext, label string) string {
	value := ctx.Value(label)
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

func SetAIModelInputToken(ctx eocontext.EoContext, token int) {
	ctx.WithValue(AIModelInputTokenLabel, token)
}

func GetAIModelInputToken(ctx eocontext.EoContext) int {
	return valueInt(ctx, AIModelInputTokenLabel)
}

func SetAIModelOutputToken(ctx eocontext.EoContext, token int) {
	ctx.WithValue(AIModelOutputTokenLabel, token)
}

func GetAIModelOutputToken(ctx eocontext.EoContext) int {
	return valueInt(ctx, AIModelOutputTokenLabel)
}

func SetAIModelTotalToken(ctx eocontext.EoContext, token int) {
	ctx.WithValue(AIModelTotalTokenLabel, token)
}

func GetAIModelTotalToken(ctx eocontext.EoContext) int {
	return valueInt(ctx, AIModelTotalTokenLabel)
}

func SetAIModelMode(ctx eocontext.EoContext, mode string) {
	ctx.WithValue(AIModelModeLabel, mode)
}

func GetAIModelMode(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelModeLabel)
}

func SetAIModel(ctx eocontext.EoContext, model string) {
	ctx.WithValue(AIModelLabel, model)
}

func GetAIModel(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelLabel)
}

func SetAIProvider(ctx eocontext.EoContext, provider string) {
	ctx.WithValue(AIProviderLabel, provider)
}

func GetAIProvider(ctx eocontext.EoContext) string {
	return valueString(ctx, AIProviderLabel)
}
