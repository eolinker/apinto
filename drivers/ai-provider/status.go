package ai_provider

import "github.com/eolinker/eosc/eocontext"

var (
	StatusNormal         = "normal"
	StatusInvalidRequest = "invalid request"
	StatusQuotaExhausted = "quota exhausted"
	StatusExpired        = "expired"
	StatusExceeded       = "exceeded"
	StatusInvalid        = "invalid"
)

func SetAIStatusExpired(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusExpired)
}

func GetAIStatusExpired(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatusQuotaExhausted(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusQuotaExhausted)
}

func GetAIStatusQuotaExhausted(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatusExceeded(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusExceeded)
}

func GetAIStatusExceeded(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatusInvalid(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusInvalid)
}

func GetAIStatusInvalid(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatusNormal(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusNormal)
}

func GetAIStatusNormal(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatusInvalidRequest(ctx eocontext.EoContext) {
	ctx.WithValue(AIModelStatusLabel, StatusInvalidRequest)
}

func GetAIStatusInvalidRequest(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}

func SetAIStatus(ctx eocontext.EoContext, status string) {
	ctx.WithValue(AIModelStatusLabel, status)
}

func GetAIStatus(ctx eocontext.EoContext) string {
	return valueString(ctx, AIModelStatusLabel)
}
