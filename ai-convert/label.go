package ai_convert

import "github.com/eolinker/eosc/eocontext"

// AI 模型转发链路上下文标签，仅用于 ai-convert/ai-provider 等 AI 链路模块。
// 计费相关标签已迁移至 github.com/eolinker/apinto/pricing 包，请使用通用计费标签。
var (
	AIModelInputTokenLabel  = "ai_model_input_token"  // 输入 token 数量
	AIModelOutputTokenLabel = "ai_model_output_token" // 输出 token 数量
	AIModelTotalTokenLabel  = "ai_model_total_token"  // 总 token 数量
	AIModelModeLabel        = "ai_model_mode"         // 模型计费模式
	AIModelLabel            = "ai_model"              // 模型名称
	AIKeyLabel              = "ai_key"                // 使用的 AI Key ID
	AIProviderLabel         = "ai_provider"           // 供应商名称
	AIProviderStatusesLabel = "ai_provider_statuses"  // 供应商计费状态列表
	AIModelStatusLabel      = "ai_model_status"       // 模型状态
)

// valueInt 从上下文中读取 int 值
func valueInt(ctx eocontext.EoContext, label string) int {
	value := ctx.Value(label)
	if v, ok := value.(int); ok {
		return v
	}
	return 0
}

// valueString 从上下文中读取 string 值
func valueString(ctx eocontext.EoContext, label string) string {
	value := ctx.Value(label)
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

// --- AI 模型 Token 标签存取函数 ---

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

func SetAIKey(ctx eocontext.EoContext, key string) {
	ctx.WithValue(AIKeyLabel, key)
}

func GetAIKey(ctx eocontext.EoContext) string {
	return valueString(ctx, AIKeyLabel)
}

func SetAIProvider(ctx eocontext.EoContext, provider string) {
	ctx.WithValue(AIProviderLabel, provider)
}

func GetAIProvider(ctx eocontext.EoContext) string {
	return valueString(ctx, AIProviderLabel)
}

// AIProviderStatus 供应商计费状态记录
type AIProviderStatus struct {
	Provider string `json:"provider"` // 供应商名称
	Model    string `json:"model"`    // 模型名称
	Key      string `json:"key"`      // 使用的 AI Key ID
	Status   string `json:"status"`   // 状态（如 quota_exhausted、exceeded）
}

// GetAIProviderStatuses 获取供应商计费状态列表
func GetAIProviderStatuses(ctx eocontext.EoContext) []AIProviderStatus {
	tmp := ctx.Value(AIProviderStatusesLabel)
	statuses := make([]AIProviderStatus, 0)
	if tmp != nil {
		result, ok := tmp.([]AIProviderStatus)
		if ok {
			statuses = result
		}
	}
	return statuses
}

// SetAIProviderStatuses 追加一条供应商计费状态记录到上下文
func SetAIProviderStatuses(ctx eocontext.EoContext, status string) {
	providerStatus := AIProviderStatus{
		Provider: GetAIProvider(ctx),
		Model:    GetAIModel(ctx),
		Key:      GetAIKey(ctx),
		Status:   status,
	}
	tmp := ctx.Value(AIProviderStatusesLabel)
	statuses := make([]AIProviderStatus, 0)
	if tmp != nil {
		result, ok := tmp.([]AIProviderStatus)
		if ok {
			statuses = result
		}
	}
	statuses = append(statuses, providerStatus)
	ctx.WithValue(AIProviderStatusesLabel, statuses)
}
