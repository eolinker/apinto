package dynamic_billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"net/http"
	"strconv"
	"strings"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/encoder"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/redis/go-redis/v9"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ http_context.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	redisID                 string
	defaultConcurrencyLimit int
	enableBalance           bool
	taskKeyGenerator        context_label.IKeyGenerator
	concurrencyKeyGenerator context_label.IKeyGenerator
	balanceKeyGenerator     context_label.IKeyGenerator
	priceKeyGenerator       context_label.IKeyGenerator
}

// TaskInfo 描述了异步任务生成的元数据，用于二次状态查询时反查账户和资源组
type TaskInfo struct {
	ResourceID  string                      `json:"resource"`
	App         string                      `json:"app"`
	IsCharged   bool                        `json:"is_charged"`
	Variables   map[string]interface{}      `json:"variables,omitempty"`
	PricingData *price_calcular.PricingData `json:"pricing_data,omitempty"`
	Cache       struct {
		Body string `json:"body"`
	} `json:"cache"`
}

func (e *executor) Reset(conf interface{}, wks map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigType
	}
	return e.reset(cfg, wks)
}

func (e *executor) reset(cfg *Config, wks map[eosc.RequireId]eosc.IWorker) error {
	e.redisID = string(cfg.Cache)

	e.defaultConcurrencyLimit = cfg.ConcurrencyLimit
	e.enableBalance = cfg.EnableBalance
	e.balanceKeyGenerator = context_label.NewKeyGenerator(cfg.BalanceKey)
	e.priceKeyGenerator = context_label.NewKeyGenerator(cfg.PriceKey)
	e.taskKeyGenerator = context_label.NewKeyGenerator(cfg.TaskKey)
	e.concurrencyKeyGenerator = context_label.NewKeyGenerator(cfg.ConcurrencyKey)
	return nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	app := ctx.GetLabel("application")
	api := ctx.GetLabel("api")
	if app == "" || api == "" {
		return nil
	}
	var err error

	// ==========================================
	// 1. 动态从 scope_manager 检索 Redis 缓存
	// ==========================================
	var cache resources.ICache
	var cl []resources.ICache
	if e.redisID != "" {
		cl = scope_manager.Get[resources.ICache](e.redisID).List()
	}
	if len(cl) == 0 {
		cl = scope_manager.Get[resources.ICache]("redis").List()
	}
	if len(cl) > 0 {
		cache = cl[0]
	} else {
		cache = resources.LocalCache()
	}

	// ==========================================
	// 2. 解析多层计费流属性 (Billing Mode, Step, Task ID)
	// ==========================================
	billingMode := context_label.GetBillingMode(ctx)
	if billingMode == "" {
		billingMode = context_label.BillingModeImmediate
	}

	resourceID := ctx.GetLabel("resource")

	// ==========================================
	// 3. 异步两阶段“查询结果 (query)”步骤的前置上下文反查还原与直接缓存拦截
	// ==========================================
	var snapshottedPricingData *price_calcular.PricingData
	var isCharged bool
	if billingMode == context_label.BillingModeTaskQuery && cache != nil {
		taskInfoKey := e.taskKeyGenerator.Key(ctx)
		infoStr, rErr := cache.Get(ctx.Context(), taskInfoKey).Result()
		if rErr == nil {
			var info TaskInfo
			if json.Unmarshal([]byte(infoStr), &info) == nil {
				// 高靠谱还原用户、资源和应用标识（免除无鉴权查询时的账户漏刷）
				if info.ResourceID != "" {
					resourceID = info.ResourceID
					ctx.SetLabel("resource", resourceID)
				}
				if info.App != "" {
					app = info.App
					ctx.SetLabel("application", app)
				}
				// 固化定价数据还原
				if info.PricingData != nil {
					snapshottedPricingData = info.PricingData
				}
				context_label.SetPriceVariables(ctx, info.Variables)

				isCharged = info.IsCharged
				// 若已有成功计费并落盘缓存的响应，则直接写回响应体阻断拦截，防范重复调用与二次扣费
				if isCharged && info.Cache.Body != "" {
					context_label.SetAmountCost(ctx, "0")
					context_label.SetAmountSale(ctx, "0")
					context_label.SetAmountOfficial(ctx, "0")
					ctx.Response().SetStatus(http.StatusOK, "200")
					ctx.Response().SetHeader("Content-Type", "application/json")
					ctx.Response().SetBody([]byte(info.Cache.Body))
					//calc, ok := w.Calculator().(*pricing_policy.Calculator)
					//_, calcErr := calc.Calculate(ctx, e.enableBalance, info.PricingData, e.extenderKeys...)
					//if calcErr != nil {
					//	log.Errorf("[dynamic-billing] calculate failed for resource %s: %v", resourceID, calcErr)
					//	ctx.SetLabel("pricing_status", "failed")
					//	ctx.SetLabel("pricing_error", calcErr.Error())
					//	return calcErr
					//}
					return nil
				}
			}
		}
	}

	if resourceID == "" {
		log.Debug("[dynamic-billing] resource_id not found, skipping pricing calculation.")
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	id := fmt.Sprintf("%s:%s", ctx.GetLabel("resource_type"), resourceID)
	//// 4. 定位计费计算器
	//w, has := policyManager.Get(id)
	//if !has {
	//	if next != nil {
	//		return next.DoChain(ctx)
	//	}
	//	return nil
	//}

	calc, ok := price_calcular.GetCalculator(id)
	if !ok || calc == nil {
		log.Errorf("[dynamic-billing] pricing-policy calculator is nil or invalid for worker %s", resourceID)
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	// ==========================================
	// 5. 并发限制检查 (Concurrency Check)
	// ==========================================
	concurrencyLimit := e.defaultConcurrencyLimit
	if limitStr := ctx.GetLabel("concurrency_limit"); limitStr != "" {
		if l, parseErr := strconv.Atoi(limitStr); parseErr == nil {
			concurrencyLimit = l
		}
	} else if limitStr := ctx.GetLabel("resource_concurrency_limit"); limitStr != "" {
		if l, parseErr := strconv.Atoi(limitStr); parseErr == nil {
			concurrencyLimit = l
		}
	}

	concurrencyKey := e.concurrencyKeyGenerator.Key(ctx)
	var acquired bool
	if concurrencyLimit > 0 && cache != nil {
		resCmd := cache.IncrBy(ctx.Context(), concurrencyKey, 1, 60*time.Second)
		val, cmdErr := resCmd.Result()
		if cmdErr == nil {
			acquired = true
			if int(val) > concurrencyLimit {
				_ = cache.DecrBy(ctx.Context(), concurrencyKey, 1, 60*time.Second)
				log.Errorf("[resource-pricing] concurrency limit exceeded for user %s on resource %s: current=%d, limit=%d", app, resourceID, val, concurrencyLimit)
				ctx.Response().SetStatus(http.StatusTooManyRequests, "429")
				ctx.Response().SetBody([]byte(`{"error":"concurrency limit exceeded"}`))
				return nil
			}
		}
	}

	if acquired {
		defer func() {
			_ = cache.DecrBy(ctx.Context(), concurrencyKey, 1, 60*time.Second)
		}()
	}
	// ==========================================
	// 10. 组装资源定价价格 Key (优先使用固化的 PricingData，若无则从 Redis 动态读取)
	// ==========================================
	var priceData price_calcular.PricingData
	if e.enableBalance {
		if snapshottedPricingData != nil {
			priceData = *snapshottedPricingData
		} else {
			priceKey := e.priceKeyGenerator.Key(ctx)

			strResult := cache.Get(ctx.Context(), priceKey)
			val, err := strResult.Result()
			if err != nil {
				errStr := fmt.Sprintf("{\"error\":\"pricing(%s) data not found\"}", priceKey)
				ctx.Response().SetBody([]byte(errStr))
				ctx.Response().SetStatus(400, "Bad Request")
				log.Errorf("[dynamic-billing] get redis price for key %s error: %v", priceKey, err)
				return err
			}

			if err := json.Unmarshal([]byte(val), &priceData); err != nil {
				errStr := fmt.Sprintf("{\"error\":\"invalid pricing(%s) data\"}", priceKey)
				ctx.Response().SetBody([]byte(errStr))
				ctx.Response().SetStatus(400, "Bad Request")
				log.Errorf("[dynamic-billing] unmarshal redis pricing data error: %v, raw data: %s", err, val)
				return err
			}
			if priceData.BasicInfo != nil {
				context_label.SetPriceVersion(ctx, priceData.BasicInfo.Version)
				context_label.SetPriceVersion(ctx, priceData.BasicInfo.Rely)
			}
		}
	}

	// ==========================================
	// 6. 组装余额扣减 Key
	// ==========================================
	balanceKey := e.balanceKeyGenerator.Key(ctx)

	// ==========================================
	// 7. 余额前置阻断校验 (Balance Pre-check)
	//    若定价数据中存在任一"免费策略"（PricePlan 的 Sale map 全为 0），
	//    则视为该资源存在免费通路，跳过余额拦截；否则按余额<=0 拦截。
	// ==========================================
	isPreCheckRequired := billingMode == context_label.BillingModeImmediate || billingMode == context_label.BillingModeTaskCreate
	if isPreCheckRequired && e.enableBalance && cache != nil && !hasFreePricePlan(&priceData) {
		balanceStr, err := cache.Get(ctx.Context(), balanceKey).Result()
		if err == nil {
			if balanceInt, parseErr := strconv.ParseInt(balanceStr, 10, 64); parseErr == nil {
				if balanceInt <= 0 {
					balance := float64(balanceInt) / 100000.0
					log.Errorf("[resource-pricing] insufficient balance for user %s: %f", app, balance)
					err = errors.New(`{"error":"insufficient balance"}`)
					ctx.Response().SetStatus(http.StatusPaymentRequired, "402")
					ctx.Response().SetBody([]byte(err.Error()))
					return err
				}
			}
		} else {
			if errors.Is(err, redis.Nil) {
				log.Warnf("[resource-pricing] balance key not found for user %s, treating as zero balance", app)
				ctx.Response().SetStatus(http.StatusPaymentRequired, "402")
				ctx.Response().SetBody([]byte(`{"error":"insufficient balance"}`))
				return err
			}
			log.Errorf("[resource-pricing] error fetching balance for user %s: %v", app, err)
			return err
		}
	}

	fn := ctx.Proxy().GetStreamBodyParse()
	//var res *pricing_policy.CalculateResult
	if ctx.GetLabel("resource_type") == "api" {
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			res, err := calc.Calculate(ctx, e.enableBalance, &priceData)
			if err != nil {
				log.Errorf("[dynamic-billing] calculate error: %v", err)
				return
			}
			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if e.enableBalance && cache != nil && res.Sale > 0 {
				success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
				if !success {
					return
				}
			}
		})
	} else if ctx.GetLabel("resource_type") == "ai" {
		var res *price_calcular.CalculateResult
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			fn := context_label.GetResponseChunkFunc(ctx)
			if fn == nil {
				if e.enableBalance && cache != nil && res != nil && res.Sale > 0 {
					success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
					if !success {
						log.Errorf("[dynamic-billing] balance deduction failed for user %s, sale amount: %f", balanceKey, res.Sale)
						return
					}
				}
				return
			}
			body, err := fn(ctx)
			if err != nil {
				log.Errorf("[dynamic-billing] get response chunk error: %v", err)
				return
			}
			res, err = calc.CalculateFromChunk(ctx, e.enableBalance, body, &priceData)
			if err != nil {
				log.Errorf("[dynamic-billing] calculate error: %v", err)
				return
			}
			if res == nil {
				return
			}
			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if e.enableBalance && cache != nil && res.Sale > 0 {
				success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
				if !success {
					return
				}
			}
			return
		})
		// 只有文本模型需要异步
		ctx.Proxy().AppendStreamBodyHandle(func(ctx http_context.IHttpContext, p []byte) ([]byte, error) {
			var body []byte
			if fn != nil {
				body = fn(ctx, p)
			} else {
				body = p
			}
			encoding := ctx.Response().Headers().Get("content-encoding")
			if encoding != "utf-8" && encoding != "" {
				body, err = encoder.ToUTF8(encoding, body)
				if err != nil {
					log.Errorf("[dynamic-billing] failed to convert response body to UTF-8: %v", err)
					return p, nil
				}
			}
			// # 遍历，以换行符分行
			lines := strings.Split(string(body), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				tmp, err := calc.CalculateFromChunk(ctx, e.enableBalance, []byte(line), &priceData)
				if err != nil {
					continue
				}
				if tmp != nil {
					res = tmp
				}
			}
			if res == nil {
				return p, nil
			}
			if res.Cost > 0 {
				context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			}
			if res.Sale > 0 {
				context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			}
			if res.Official > 0 {
				context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			}

			return p, nil
		})
	}

	// ==========================================
	// 8. 转发下游链路 (Do Chain Forwarding)
	// ==========================================
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}

	// ==========================================
	// 9. 异步两阶段“生成任务 (create)”后置落盘保存
	// ==========================================
	if context_label.IsBillingMode(ctx, context_label.BillingModeTaskCreate) {
		fn := context_label.GetTaskIDSetFunc(ctx)
		if fn != nil {
			err := fn(ctx)
			if err != nil {
				return err
			}
		}
		extractor := calc.VariablesExtractor()
		if extractor == nil {
			calculator := price_calcular.GetICalculator(ctx)
			if calculator == nil {
				log.Errorf("[dynamic-billing] calculator is nil, cannot extract variables for task info")
				ctx.Response().SetBody([]byte("{\"error\":\"calculator is nil\"}"))
				ctx.Response().SetStatus(500, "Internal Server Error")
				return fmt.Errorf("calculator is nil")
			}
			extractor = calculator.VariablesExtractor()
		}
		taskInfoKey := e.taskKeyGenerator.Key(ctx)
		info := TaskInfo{
			ResourceID: resourceID,
			App:        app,
			Variables:  extractor.ExtractAll(ctx),
		}

		// 获取并固化当前的定价配置数据，确保 query 阶段计费的一致性，防止中途价格配置变更
		priceKey := e.priceKeyGenerator.Key(ctx)
		if priceVal, pErr := cache.Get(ctx.Context(), priceKey).Result(); pErr == nil {
			var rData price_calcular.PricingData
			if json.Unmarshal([]byte(priceVal), &rData) == nil {
				info.PricingData = &rData
			}
		}

		if infoBytes, mErr := json.Marshal(info); mErr == nil {
			_ = cache.Set(ctx.Context(), taskInfoKey, infoBytes, 24*time.Hour)
			log.DebugF("[dynamic-billing] [create_task] successfully cached async task metadata and pricing snapshot for %s", taskInfoKey)
		}

		return nil
	}

	if ctx.Response().IsBodyStream() {
		return nil
	}

	// 同步结算非流式响应
	return e.settleBilling(ctx, calc, &priceData, balanceKey, resourceID, app, cache, billingMode, isCharged)
}

func (e *executor) immediateSettle(ctx http_context.IHttpContext, calc price_calcular.ICalculator, priceData *price_calcular.PricingData, balanceKey, resourceID, app string, cache resources.ICache) error {
	var res *price_calcular.CalculateResult
	var err error
	res, err = calc.Calculate(ctx, e.enableBalance, priceData)
	if err != nil {
		log.Errorf("[dynamic-billing] calculate failed for resource %s: %v", resourceID, err)
		return err
	}

	context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
	context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
	context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))

	if e.enableBalance && cache != nil && res.Sale > 0 {
		success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
		if !success {
			return fmt.Errorf("insufficient balance during settlement")
		}
	}
	return nil
}

func (e *executor) settleBilling(ctx http_context.IHttpContext, calc price_calcular.ICalculator, priceData *price_calcular.PricingData, balanceKey, resourceID, app string, cache resources.ICache, billingMode context_label.BillingMode, isCharged bool) error {
	switch billingMode {
	case context_label.BillingModeImmediate:
		return e.immediateSettle(ctx, calc, priceData, balanceKey, resourceID, app, cache)
	case context_label.BillingModeTaskQuery:
		if isCharged {
			// 已经计费过或者任务完成状态为false，不做重复计费
			return nil
		}
		fn := context_label.GetTaskStatusParseFunc(ctx)
		if fn == nil {
			return nil
		}
		taskStatus, err := fn(ctx)
		if err != nil {
			return err
		}
		switch taskStatus {
		case context_label.TaskStatusRunning:
		case context_label.TaskStatusSuccess:
			isCharged = true
			res, calcErr := calc.Calculate(ctx, e.enableBalance, priceData)
			if calcErr != nil {
				log.Errorf("[dynamic-billing] calculate failed for resource %s: %v", resourceID, calcErr)
				ctx.SetLabel("pricing_status", "failed")
				ctx.SetLabel("pricing_error", calcErr.Error())
				return calcErr
			}

			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if e.enableBalance && cache != nil && res.Cost > 0 {
				success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
				if !success {
					return fmt.Errorf("insufficient balance during settlement")
				}
			}
		case context_label.TaskStatusFailed:
			isCharged = true
		}
		taskKey := e.taskKeyGenerator.Key(ctx)
		taskInfo := TaskInfo{
			ResourceID: resourceID,
			App:        app,
			IsCharged:  isCharged,
			Cache: struct {
				Body string `json:"body"`
			}(struct{ Body string }{
				Body: string(ctx.Response().GetBody()),
			}),
			Variables: context_label.GetPriceVariables(ctx),
		}
		taskByte, _ := json.Marshal(taskInfo)
		cache.Set(ctx.Context(), taskKey, taskByte, 24*time.Hour)

	case context_label.BillingModeTaskCreate:

	}

	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

func (e *executor) Destroy() {
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Stop() error {
	return nil
}

// executeBalanceDeduction 在给定的 Redis 缓存上原子性扣减余额。
//
// 使用 Redis 原生 DECRBY（单命令原子、64 位整数运算）替代此前的
// "get -> 浮点相减 -> set" 方案：后者在 Redis Lua 中以 IEEE 754 双精度浮点
// 计算并回写，大整数会被格式化为科学计数法（如 4e+06），多次读改写后精度
// 逐步损坏，最终可能被写成 0。DECRBY 全程整数运算，不产生浮点/科学计数法，
// 且天然允许结果为负值。
func executeBalanceDeduction(ctx context.Context, cache resources.ICache, balanceKey string, cost float64, userID string) bool {
	// 余额与扣减金额统一以「元 × 100000」的整数存储（与前置校验的 /100000.0 保持一致），
	// 保留 5 位小数精度。先四舍五入消除浮点误差后转为整数，避免向 Redis 传入浮点。
	deductLua := `
		local balanceKey = KEYS[1]
		local deductAmount = tonumber(ARGV[1])
		return redis.call('decrby', balanceKey, deductAmount)
	`
	total := int64(math.Round(cost * 1000000))
	evalResCmd := cache.Run(ctx, deductLua, []string{balanceKey}, total)
	newBalance, evalErr := evalResCmd.Result()
	if evalErr != nil {
		log.Errorf("[resource-pricing] balance deduct redis error for user %s, cost=%f: %v", userID, cost, evalErr)
		return false
	}

	// 扣款成功
	log.DebugF("[resource-pricing] balance deducted successfully for user %s, cost=%f, key=%s, new_balance=%v", userID, cost, balanceKey, newBalance)
	return true
}

// hasFreePricePlan 判断给定的定价数据中是否存在"免费策略"：
// 只要 pricingData.Strategy 中任一 PricePlan 的 Sale map 全部为 0（且非空），
// 就认为该资源存在免费通路，前置阶段不应因余额不足而拦截。
func hasFreePricePlan(pricingData *price_calcular.PricingData) bool {
	if pricingData == nil || len(pricingData.Strategy) == 0 {
		return false
	}
	for _, plan := range pricingData.Strategy {
		if plan == nil || len(plan.Sale) == 0 {
			continue
		}
		allZero := true
		for _, v := range plan.Sale {
			if v != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			return true
		}
	}
	return false
}
