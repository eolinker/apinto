package pricing

import "github.com/eolinker/eosc/eocontext"

// 通用计费上下文标签键。值会同时通过 ctx.WithValue 写入对象、
// 通过 ctx.SetLabel 写入 label，便于其他插件、日志、计费下游消费。
const (
	PricingResultLabel   = "pricing_result"   // 计费计算结果（*CalculateResult）
	PricingFieldsLabel   = "pricing_fields"   // 计费字段集合（map[string]interface{}）
	PricingPhaseLabel    = "pricing_phase"    // 计费阶段（PricingPhase）
	PricingAccountLabel  = "pricing_account"  // 账户 ID
	PricingProviderLabel = "pricing_provider" // 供应商标识（供拦截插件查找计算器兼底）
	PricingResourceLabel = "pricing_resource" // 资源标识（AI 场景=模型名，API 场景=接口名）
)

// SetPricingResult 写入计费结果到上下文。
func SetPricingResult(ctx eocontext.EoContext, result *CalculateResult) {
	ctx.WithValue(PricingResultLabel, result)
}

// GetPricingResult 读取计费结果。
func GetPricingResult(ctx eocontext.EoContext) *CalculateResult {
	if v, ok := ctx.Value(PricingResultLabel).(*CalculateResult); ok {
		return v
	}
	return nil
}

// SetPricingFields 写入计费字段集合到上下文。
func SetPricingFields(ctx eocontext.EoContext, fields map[string]interface{}) {
	ctx.WithValue(PricingFieldsLabel, fields)
}

// GetPricingFields 读取计费字段集合。
func GetPricingFields(ctx eocontext.EoContext) map[string]interface{} {
	if v, ok := ctx.Value(PricingFieldsLabel).(map[string]interface{}); ok {
		return v
	}
	return nil
}

// SetPricingPhase 写入计费阶段。
func SetPricingPhase(ctx eocontext.EoContext, phase PricingPhase) {
	ctx.WithValue(PricingPhaseLabel, phase)
}

// GetPricingPhase 读取计费阶段，未设置时返回同步阶段。
func GetPricingPhase(ctx eocontext.EoContext) PricingPhase {
	if v, ok := ctx.Value(PricingPhaseLabel).(PricingPhase); ok {
		return v
	}
	return PricingPhaseSync
}

// GetAccountID 从上下文 label 中读取账户 ID。
// 仅识别通用 PricingAccountLabel；不再保留旧的 ai_consumer 别名以维持抽象干净。
func GetAccountID(ctx eocontext.EoContext) string {
	return ctx.GetLabel(PricingAccountLabel)
}

// GetProvider 从上下文 label 中读取供应商标识。
// AI 网关如需对接，可在前置流程将原 ai_provider 逆写到 PricingProviderLabel。
func GetProvider(ctx eocontext.EoContext) string {
	return ctx.GetLabel(PricingProviderLabel)
}

// GetResource 从上下文 label 中读取资源标识。
func GetResource(ctx eocontext.EoContext) string {
	return ctx.GetLabel(PricingResourceLabel)
}
