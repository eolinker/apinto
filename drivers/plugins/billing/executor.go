package billing

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// customExtractorCache 缓存计算器自定义字段提取器，避免每次请求重复编译 JSONPath。
// 键采用 IPriceCalculator 实例指针，计算器变更时旧条目自动失效。
var customExtractorCache sync.Map // map[pricing.IPriceCalculator]*FieldExtractor

// getCustomExtractor 获取或编译计算器级字段提取器；
// 若计算器既未声明 RequestFields 又未声明 ResponseFields 则返回 nil。
func getCustomExtractor(calc pricing.IPriceCalculator) (*FieldExtractor, error) {
	if calc == nil {
		return nil, nil
	}
	reqFields := calc.RequestFields()
	respFields := calc.ResponseFields()
	if len(reqFields) == 0 && len(respFields) == 0 {
		return nil, nil
	}
	if val, ok := customExtractorCache.Load(calc); ok {
		return val.(*FieldExtractor), nil
	}
	extractor, err := NewFieldExtractor(reqFields, respFields)
	if err != nil {
		return nil, err
	}
	customExtractorCache.Store(calc, extractor)
	return extractor, nil
}

var _ eosc.IWorker = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ http_context.HttpFilter = (*executor)(nil)

// executor 是 billing 插件的核心，承担「拦截 → 提取 → 查找计算器 → 预扣 → 转发 → 计算 → 结算」全流程。
type executor struct {
	drivers.WorkerBase
	provider       string
	resource       string
	phase          pricing.PricingPhase
	extractor      *FieldExtractor
	matchRules     *matchRulesMatcher
	resultMatcher  *resultMatcher
	taskExecutor   pricing.ITaskExecutor
	balanceManager pricing.IBalanceManager
}

func (e *executor) Start() error { return nil }

// Reset 在配置变更时重建依赖的字段提取器与匹配器，复用 worker 实例避免抖动。
func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}

	extractor, err := NewFieldExtractor(cfg.RequestFields, cfg.ResponseFields)
	if err != nil {
		return fmt.Errorf("create field extractor: %w", err)
	}

	var matcher *matchRulesMatcher
	if cfg.MatchRules != nil {
		matcher, err = newMatchRulesMatcher(cfg.MatchRules)
		if err != nil {
			return fmt.Errorf("create match rules: %w", err)
		}
	}

	var rMatcher *resultMatcher
	if cfg.ResultMatch != nil {
		rMatcher, err = newResultMatcher(cfg.ResultMatch, extractor)
		if err != nil {
			return fmt.Errorf("create result matcher: %w", err)
		}
	}

	e.provider = cfg.Provider
	e.resource = cfg.Resource
	e.phase = cfg.Phase
	e.extractor = extractor
	e.matchRules = matcher
	e.resultMatcher = rMatcher
	return nil
}

// Stop 停止异步任务执行器并清空匹配器/提取器引用，便于 GC。
func (e *executor) Stop() error {
	if e.taskExecutor != nil {
		e.taskExecutor.Stop()
	}
	e.extractor = nil
	e.matchRules = nil
	e.resultMatcher = nil
	return nil
}

func (e *executor) Destroy() {
	if e.taskExecutor != nil {
		e.taskExecutor.Stop()
	}
	e.extractor = nil
	e.matchRules = nil
	e.resultMatcher = nil
	e.taskExecutor = nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {
	return http_context.DoHttpFilter(e, ctx, next)
}

// DoHttpFilter 计费拦截主流程：
//
//  1. 请求匹配（matchRules）：未命中则直接放行
//  2. 请求字段提取
//  3. 确定 provider/resource/phase（请求字段 → 上下文 label → executor 默认）
//  4. 查找计算器（带 phase 与去 phase 两级降级）
//  5. 计算器自定义字段补充提取
//  6. PreDeduct 预扣（可选）：余额不足拦截 402，其他错误降级放行
//  7. 转发下游
//  8. 失败时 Rollback
//  9. 响应字段提取并合并
//  10. 异步任务侦测：仅返回 task_id 无 usage 时切换为 submit 阶段
//  11. 已计费幂等：task_id 已标记则跳过重复计费
//  12. 结果匹配：不满足则跳过 Calculate
//  13. Calculate + MarkBilled
//  14. Settle 实际费用 / 写入 ctx label
func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	// 1. 请求匹配
	if e.matchRules != nil && !e.matchRules.Match(ctx) {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	// 2. 请求字段提取
	reqBody, _ := ctx.Proxy().Body().RawBody()
	reqFields, err := e.extractor.ExtractFromRequest(reqBody)
	if err != nil {
		log.Printf("[billing] extract request fields error: %v", err)
		reqFields = map[string]interface{}{}
	}

	// 3. 确定 provider/resource
	provider := determineProvider(reqFields, ctx, e.provider)
	resource := determineResource(reqFields, ctx, e.resource)

	// 4. 查找计算器，按 phase → 空 phase 顺序两级降级
	lookupPhase := e.phase
	calculator, has := pricing.GetCalculator(provider, resource, string(lookupPhase))
	if !has && lookupPhase != "" {
		calculator, has = pricing.GetCalculator(provider, resource, "")
	}
	if !has {
		// 兜底再尝试 executor 默认 provider/resource
		calculator, has = pricing.GetCalculator(e.provider, e.resource, string(lookupPhase))
		if !has && lookupPhase != "" {
			calculator, has = pricing.GetCalculator(e.provider, e.resource, "")
		}
		if has {
			provider = e.provider
			resource = e.resource
		}
	}
	if !has {
		log.Printf("[billing] no calculator found for provider=%s resource=%s phase=%s, skip pricing", provider, resource, lookupPhase)
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	effectivePhase := calculator.Phase()
	if effectivePhase == "" {
		effectivePhase = lookupPhase
	}

	chargeType := pricing.ChargeTypeActual
	if effectivePhase == pricing.PricingPhaseSubmit {
		chargeType = pricing.ChargeTypePre
	}

	log.Printf("[billing] matched calculator: provider=%s resource=%s phase=%s charge_type=%s mode=%s",
		provider, resource, effectivePhase, chargeType, calculator.Pricing().Mode)

	// 5. 计算器自定义字段提取
	customExtractor, cerr := getCustomExtractor(calculator)
	if cerr != nil {
		log.Printf("[billing] compile custom extractor error: %v", cerr)
	}
	if customExtractor != nil {
		if customReqFields, ferr := customExtractor.ExtractFromRequest(reqBody); ferr == nil && len(customReqFields) > 0 {
			reqFields = mergeFields(reqFields, customReqFields)
		}
	}

	// 6. PreDeduct
	accountID := pricing.GetAccountID(ctx)
	var freezeID string
	var preDeductCalled bool
	if e.balanceManager != nil && accountID != "" {
		estimatedCost := estimateMaxCost(calculator)
		meta := pricing.DeductMeta{Provider: provider, Resource: resource, Phase: effectivePhase}
		fID, perr := e.balanceManager.PreDeduct(ctx.Context(), accountID, estimatedCost, meta)
		if errors.Is(perr, pricing.ErrInsufficientBalance) {
			log.Printf("[billing] insufficient balance for account=%s, intercepting request", accountID)
			ctx.Response().SetStatus(402, "402")
			ctx.Response().SetBody([]byte(`{"error":"insufficient balance"}`))
			return nil
		} else if perr != nil {
			log.Printf("[billing] pre-deduct error (bypass): %v", perr)
		} else {
			freezeID = fID
			preDeductCalled = true
			ctx.SetLabel("pricing_freeze_id", freezeID)
		}
	}

	// 7. 转发下游
	if next != nil {
		err = next.DoChain(ctx)
	}

	// 8. 失败回滚
	if err != nil && preDeductCalled && e.balanceManager != nil {
		log.Printf("[billing] upstream request failed, rolling back freezeID=%s", freezeID)
		_ = e.balanceManager.Rollback(ctx.Context(), freezeID)
		return err
	}

	// 9. 响应字段提取
	respBody := ctx.Response().GetBody()
	respFields, rerr := e.extractor.ExtractFromResponse(respBody)
	if rerr != nil {
		log.Printf("[billing] extract response fields error: %v", rerr)
		respFields = map[string]interface{}{}
	}
	if customExtractor != nil {
		if customRespFields, cerr := customExtractor.ExtractFromResponse(respBody); cerr == nil && len(customRespFields) > 0 {
			respFields = mergeFields(respFields, customRespFields)
		}
	}

	allFields := mergeFields(reqFields, respFields)
	taskID := pricing.ExtractTaskID(allFields)

	// 10. 异步任务侦测
	if chargeType == pricing.ChargeTypeActual && detectPreChargeOnly(allFields) {
		log.Printf("[billing] response has task_id but no usage, mark as submit phase: task_id=%s", taskID)
		pricing.SetPricingFields(ctx, allFields)
		pricing.SetPricingPhase(ctx, pricing.PricingPhaseSubmit)
		fieldsJSON, _ := json.Marshal(allFields)
		ctx.SetLabel(pricing.PricingFieldsLabel, string(fieldsJSON))
		if taskID != "" && e.taskExecutor != nil {
			_ = e.taskExecutor.Submit(taskID, provider, resource, effectivePhase, nil, allFields)
		}
		return err
	}

	// 11. 已计费幂等
	if taskID != "" && e.taskExecutor != nil && e.taskExecutor.IsBilled(taskID) {
		log.Printf("[billing] task_id=%s already billed, skip duplicate billing", taskID)
		pricing.SetPricingFields(ctx, allFields)
		pricing.SetPricingPhase(ctx, pricing.PricingPhaseQuery)
		ctx.SetLabel("pricing_task_duplicate", "true")
		return err
	}

	// 12. 结果匹配
	if e.resultMatcher != nil && !e.resultMatcher.Match(ctx, allFields) {
		log.Printf("[billing] response does not match result rules, skip pricing")
		pricing.SetPricingFields(ctx, allFields)
		ctx.SetLabel("pricing_result_unmatched", "true")
		// 失败匹配时回滚预扣
		if preDeductCalled && e.balanceManager != nil {
			_ = e.balanceManager.Rollback(ctx.Context(), freezeID)
		}
		return err
	}

	// 13. Calculate
	usage := buildPricingUsage(calculator.Pricing().Mode, allFields, ctx)
	result, calcErr := calculator.Calculate(usage)
	if calcErr != nil {
		log.Printf("[billing] calculate error: %v", calcErr)
		pricing.SetPricingFields(ctx, allFields)
		if preDeductCalled && e.balanceManager != nil {
			_ = e.balanceManager.Rollback(ctx.Context(), freezeID)
		}
		return err
	}

	if taskID != "" && e.taskExecutor != nil {
		e.taskExecutor.MarkBilled(taskID)
	}

	// 14. 写入 ctx label
	pricing.SetPricingResult(ctx, result)
	pricing.SetPricingFields(ctx, allFields)
	pricing.SetPricingPhase(ctx, effectivePhase)

	if preDeductCalled && e.balanceManager != nil {
		meta := pricing.DeductMeta{Provider: provider, Resource: resource, Phase: effectivePhase, TaskID: taskID}
		if serr := e.balanceManager.Settle(ctx.Context(), freezeID, result.TotalCost, meta); serr != nil {
			log.Printf("[billing] settle error: %v", serr)
		}
	}

	fieldsJSON, _ := json.Marshal(allFields)
	ctx.SetLabel(pricing.PricingFieldsLabel, string(fieldsJSON))
	resultJSON, _ := json.Marshal(result)
	ctx.SetLabel(pricing.PricingResultLabel, string(resultJSON))

	return err
}

// CheckSkill 仅声明 HttpFilter 技能。
func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

// estimateMaxCost 按计算器配置预估单次最大费用，用于 PreDeduct 冻结金额。
// 估算偏保守，实际结算时 Settle 会按真实费用多退少补。
func estimateMaxCost(calc pricing.IPriceCalculator) float64 {
	p := calc.Pricing()
	if p == nil {
		return 0
	}
	switch p.Mode {
	case pricing.PricingModePerCall:
		return p.PerCall
	case pricing.PricingModeToken:
		// 简单按 1 万 token 估算
		return p.Input*0.01 + p.Output*0.01
	case pricing.PricingModePerSecond:
		// 保守按 60 秒估
		return p.PerSecond * 60
	default:
		return 0
	}
}

// detectPreChargeOnly 判断响应是否仅有 task_id 而无任何 usage 字段（异步预扣场景）。
// 仅识别新版字段名（input_count/output_count/total_count），不再保留旧 token 别名。
func detectPreChargeOnly(fields map[string]interface{}) bool {
	hasTaskID := false
	for _, key := range []string{"task_id", "id", "request_id"} {
		if v, ok := fields[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				hasTaskID = true
				break
			}
		}
	}
	if !hasTaskID {
		return false
	}
	for _, key := range []string{"usage", "input_count", "output_count", "total_count"} {
		if _, ok := fields[key]; ok {
			return false
		}
	}
	return true
}

// determineProvider 解析当前请求的 provider：reqFields → ctx label → executor 默认。
func determineProvider(reqFields map[string]interface{}, ctx http_context.IHttpContext, fallback string) string {
	if v, ok := reqFields["provider"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if s := pricing.GetProvider(ctx); s != "" {
		return s
	}
	return fallback
}

// determineResource 解析当前请求的 resource：reqFields[resource|model] → ctx label → executor 默认。
// 兼容 model 字段以便 AI 链路无需改造即可对接。
func determineResource(reqFields map[string]interface{}, ctx http_context.IHttpContext, fallback string) string {
	for _, key := range []string{"resource", "model"} {
		if v, ok := reqFields[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	if s := pricing.GetResource(ctx); s != "" {
		return s
	}
	return fallback
}

// mergeFields 合并请求字段与响应字段；响应字段优先覆盖。
func mergeFields(reqFields, respFields map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{}, len(reqFields)+len(respFields))
	for k, v := range reqFields {
		merged[k] = v
	}
	for k, v := range respFields {
		merged[k] = v
	}
	return merged
}

// buildPricingUsage 根据计费模式构造 PricingUsage。
// 仅识别新版字段名；TotalCount 缺省时按 InputCount+OutputCount 推导；
// Method/Path 直接从 ctx 读取以支撑高级阶梯按方法/路径定价。
func buildPricingUsage(mode pricing.PricingMode, fields map[string]interface{}, ctx http_context.IHttpContext) *pricing.PricingUsage {
	usage := &pricing.PricingUsage{
		Fields:     fields,
		CallCount:  1,
		StatusCode: ctx.Response().StatusCode(),
		Method:     ctx.Proxy().Method(),
		Path:       ctx.Proxy().URI().Path(),
	}
	usage.InputCount = intField(fields, "input_count")
	usage.OutputCount = intField(fields, "output_count")
	usage.CachedCount = intField(fields, "cached_count")
	usage.TotalCount = intField(fields, "total_count")
	usage.Duration = floatField(fields, "duration")
	usage.InputType = stringField(fields, "input_type")
	usage.Size = stringField(fields, "size")
	if cc := intField(fields, "call_count"); cc > 0 {
		usage.CallCount = cc
	}
	if usage.TotalCount == 0 {
		usage.TotalCount = usage.InputCount + usage.OutputCount
	}
	usage.Success = ctx.Response().StatusCode() >= 200 && ctx.Response().StatusCode() < 300
	return usage
}

func intField(fields map[string]interface{}, key string) int {
	v, ok := fields[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0
		}
		return int(i)
	}
	return 0
}

func floatField(fields map[string]interface{}, key string) float64 {
	v, ok := fields[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}

func stringField(fields map[string]interface{}, key string) string {
	v, ok := fields[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
