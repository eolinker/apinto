// Package pricing 定义通用计费契约层，提供与具体业务（AI / 标准 API）解耦的
// 定价模型、用量数据、计算结果、动态调价、条件匹配等核心数据结构。
//
// 本包只承载契约（接口与数据类型），不包含具体的扣减实现或拦截逻辑。
// 具体实现见 drivers/pricing（计算器 worker）与 drivers/plugins/billing（拦截插件）。
package pricing

// PricingMode 定义计费模式，覆盖 AI 与标准 API 两类常见计费场景。
type PricingMode string

const (
	// PricingModeToken 按量计费（如 token、字节数、记录条数），单价以"每百万单位"计算。
	PricingModeToken PricingMode = "token"
	// PricingModePerCall 按次计费，无论用量多少每次调用固定金额。
	PricingModePerCall PricingMode = "per_call"
	// PricingModePerSecond 按耗时计费，适用于流式接口、长任务。
	PricingModePerSecond PricingMode = "per_second"
	// PricingModeAdvanced 高级条件计费，按 AdvancedRule 的 Condition 命中阶梯定价。
	PricingModeAdvanced PricingMode = "advanced"
)

// PricingPhase 定义计费阶段，主要用于异步任务（如文生视频）的两阶段计费模型。
// 同步 API（包括大多数标准 API、聊天补全 API）一般保持空字符串即可。
type PricingPhase string

const (
	// PricingPhaseSync 同步计费，请求完成立即结算（默认）。
	PricingPhaseSync PricingPhase = ""
	// PricingPhaseSubmit 提交阶段，对异步任务在提交时按预估值预扣。
	PricingPhaseSubmit PricingPhase = "submit"
	// PricingPhaseQuery 查询阶段，异步任务结果到位后按实际用量结算。
	PricingPhaseQuery PricingPhase = "query"
)

// ChargeType 定义计费类型，与 PricingPhase 存在映射关系：
// PricingPhaseSync / PricingPhaseQuery → ChargeTypeActual；PricingPhaseSubmit → ChargeTypePre。
type ChargeType string

const (
	ChargeTypeActual ChargeType = "actual"
	ChargeTypePre    ChargeType = "pre"
)

// PriceLevel 定义价格层级，作为动态调价策略的参考基准。
type PriceLevel string

const (
	PriceLevelOfficial PriceLevel = "official"
	PriceLevelPurchase PriceLevel = "purchase"
	PriceLevelSale     PriceLevel = "sale"
)

// ConditionType 定义高级阶梯计费的匹配条件类型。
// 注：相对旧版已移除 total_tokens/video_size 等 AI 旧别名，改用通用名称；
// 新增 method/path 两类条件方便 API 场景使用。
type ConditionType string

const (
	ConditionTotalCount ConditionType = "total_count" // 按总用量匹配（数值阈值）
	ConditionInputType  ConditionType = "input_type"  // 按输入类型字符串匹配（text/image/...）
	ConditionSize       ConditionType = "size"        // 按尺寸字符串匹配（如 1080p、1024x1024）
	ConditionStatusCode ConditionType = "status_code" // 按 HTTP 状态码数值阈值匹配
	ConditionMethod     ConditionType = "method"      // 按 HTTP 方法字符串匹配
	ConditionPath       ConditionType = "path"        // 按 HTTP path 字符串匹配
	ConditionField      ConditionType = "field"       // 按 PricingUsage.Fields 自定义字段匹配
)

// PricingUsage 计费用量，作为 IPriceCalculator.Calculate 的输入。
// 字段全部使用通用命名，AI token 旧字段名（如 input_tokens）由 drivers/plugins/billing
// 在 buildPricingUsage 阶段做兼容映射。
type PricingUsage struct {
	InputCount  int                    `json:"input_count"`      // 输入用量
	OutputCount int                    `json:"output_count"`     // 输出用量
	CachedCount int                    `json:"cached_count"`     // 缓存命中用量
	TotalCount  int                    `json:"total_count"`      // 总量，缺省时由 InputCount + OutputCount 推导
	Duration    float64                `json:"duration"`         // 耗时（秒）
	CallCount   int                    `json:"call_count"`       // 调用次数，默认 1
	InputType   string                 `json:"input_type"`       // 输入类型（text/image/...）
	Size        string                 `json:"size"`             // 尺寸字符串
	StatusCode  int                    `json:"status_code"`      // HTTP 状态码
	Method      string                 `json:"method"`           // HTTP 方法
	Path        string                 `json:"path"`             // HTTP path
	Success     bool                   `json:"success"`          // 是否被判定为成功
	Fields      map[string]interface{} `json:"fields,omitempty"` // 通用扩展字段，ConditionField 从此取值
}

// CalculateResult 计费计算结果。
type CalculateResult struct {
	TotalCost    float64      `json:"total_cost"`
	InputCost    float64      `json:"input_cost"`
	OutputCost   float64      `json:"output_cost"`
	CachedCost   float64      `json:"cached_cost"`
	CallCost     float64      `json:"call_cost"`
	DurationCost float64      `json:"duration_cost"`
	Mode         PricingMode  `json:"mode"`
	Phase        PricingPhase `json:"phase"`
	ChargeType   ChargeType   `json:"charge_type"`
}

// ResourcePricing 资源定价配置。"资源"在 AI 场景指模型，在 API 场景指接口/服务标识。
// 单价语义：Input/Output/CachedInput 按"每百万单位"；PerCall/PerSecond 按"每次/每秒"直接相乘。
type ResourcePricing struct {
	Resource      string         `json:"resource"`                 // 资源名（AI 场景=模型名；API 场景=接口/服务标识）
	Mode          PricingMode    `json:"mode"`                     // 计费模式
	Input         float64        `json:"input"`                    // 输入单价（每百万单位）
	Output        float64        `json:"output"`                   // 输出单价（每百万单位）
	CachedInput   float64        `json:"cached_input"`             // 缓存命中输入单价（每百万单位）
	PerCall       float64        `json:"per_call"`                 // 每次调用单价
	PerSecond     float64        `json:"per_second"`               // 每秒单价
	AdvancedRules []AdvancedRule `json:"advanced_rules,omitempty"` // 高级阶梯规则
}

// AdvancedRule 高级计费规则；命中第一条匹配规则后按其 TierPricing 计算，
// 全部不匹配时回退使用 ResourcePricing 的基础单价。
type AdvancedRule struct {
	Condition Condition   `json:"condition"`
	Pricing   TierPricing `json:"pricing"`
}

// Condition 高级阶梯计费条件。
type Condition struct {
	Type      ConditionType `json:"type"`
	Operator  string        `json:"operator,omitempty"`  // 数值比较运算符：<=, >, >=, <, ==, =
	Threshold float64       `json:"threshold,omitempty"` // 数值阈值
	Value     string        `json:"value,omitempty"`     // 字符串期望值
	Field     string        `json:"field,omitempty"`     // 仅 Type=field 时使用
}

// TierPricing 阶梯定价。命中阶梯时替代基础单价，支持四种维度的阶梯化。
type TierPricing struct {
	Input       float64 `json:"input,omitempty"`
	Output      float64 `json:"output,omitempty"`
	CachedInput float64 `json:"cached_input,omitempty"`
	PerCall     float64 `json:"per_call,omitempty"`
	PerSecond   float64 `json:"per_second,omitempty"`
}
