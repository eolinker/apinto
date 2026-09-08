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
	// PreDeductInputTokensText 文本模型预扣输入估算 token 数（10K，即 1% 百万 token 单价）
	PreDeductInputTokensText = 10000
	// PreDeductOutputTokensText 文本模型预扣输出估算 token 数（1K，即 0.11% 百万 token 单价）
	PreDeductOutputTokensText = 1000
	// PreDeductTokensImage 图片/视频模型预扣输出估算 token 数（1M，即 100% 百万 token 单价）
	PreDeductTokensImage = 10000

	PreDeductTokensVideo = 100000
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

// MaxSaleStrategyID 获取 Sale 总值最高的 strategy 的 id。
// 若 PricingData 为空、Strategy 为空或无有效 Sale 数据，返回空字符串。
func (p *PricingData) MaxSaleStrategyID() string {
	if p == nil || len(p.Strategy) == 0 {
		return ""
	}
	var (
		maxID  string
		maxSum float64
		has    bool
	)
	for id, plan := range p.Strategy {
		if plan == nil {
			continue
		}
		var (
			sum     float64
			planHas bool
		)
		for _, v := range plan.Sale {
			sum += v
			planHas = true
		}
		if !planHas {
			continue
		}
		if !has || sum > maxSum {
			maxSum = sum
			maxID = id
			has = true
		}
	}
	return maxID
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

// MinSalePricingData 从 []*PricingData 中挑选 Sale 总和最小的一个。
// Sale 总和 = 遍历所有 Strategy 下 PricePlan.Sale 里的所有数值累加。
// 若入参为空或全部无有效 Sale 数据，返回 nil。
func MinSalePricingData(list []*PricingData) *PricingData {
	var (
		minData *PricingData
		minSum  float64
	)
	for _, data := range list {
		if data == nil {
			continue
		}
		sum, ok := sumSale(data)
		if !ok {
			continue
		}
		if minData == nil || sum < minSum {
			minData = data
			minSum = sum
		}
	}
	return minData
}

// sumSale 计算单个 PricingData 中所有 Strategy 下 Sale 的累加值。
// 返回值 ok 为 false 表示没有任何有效的 Sale 数据。
func sumSale(data *PricingData) (float64, bool) {
	var (
		sum float64
		has bool
	)
	for _, plan := range data.Strategy {
		if plan == nil {
			continue
		}
		for _, v := range plan.Sale {
			sum += v
			has = true
		}
	}
	return sum, has
}
