package price_calcular

// 计费模式常量：区分资源在计价维度上的单位差异。
const (
	// BillingModePerCall 按次计费（默认适用于 api 资源，以及部分单次调用型 AI 模型如生歌模型）
	BillingModePerCall = "per_call"
	// BillingModeToken 按 token 计费（大部分文本类 LLM）
	BillingModeToken = "token"
	// BillingModePerSecond 按秒计费（视频 / 音频类模型）
	BillingModePerSecond = "per_second"
)

// 模型类型常量：仅在 BillingModeToken 场景下用于区分预扣阶段的输出估算基数。
const (
	ModelTypeText  = "text"
	ModelTypeImage = "image"
	ModelTypeVideo = "video"
)

// 预扣阶段的默认参数
const (
	// DefaultSafetyFactor 预扣输出成本的默认安全系数
	DefaultSafetyFactor = 1.2
	// PreDeductSecondsForPerSecond 按秒计费时，预扣默认按 5 秒计价
	PreDeductSecondsForPerSecond = 5
	// PreDeductTokensText 文本模型预扣输出估算 token 数（10K，即 1% 百万 token 单价）
	PreDeductTokensText = 10000
	// PreDeductTokensImageVideo 图片/视频模型预扣输出估算 token 数（1M，即 100% 百万 token 单价）
	PreDeductTokensImageVideo = 1000000
	// PriceTokenScale token 单价除数（价格表按百万 token 单价配置）
	PriceTokenScale = 1000000

	// LabelInputToken 上下文中传递预估 input_token 的标签名
	LabelInputToken = "input_token"
	// LabelModelType 上下文中传递模型类型的标签名（text/image/video）
	LabelModelType = "model_type"
)

type PricingData struct {
	BasicInfo *BasicInfo            `json:"basic_info"`
	Strategy  map[string]*PricePlan `json:"strategy"`
}

type BasicInfo struct {
	Version         string `json:"version"`
	Rely            string `json:"rely"`
	ResourceGroupID string `json:"resource_group_id"`
	TenantID        string `json:"tenant_id"`
}

type PricePlan struct {
	Cost     map[string]float64 `json:"cost"`
	Sale     map[string]float64 `json:"sale"`
	Official map[string]float64 `json:"official"`
}
