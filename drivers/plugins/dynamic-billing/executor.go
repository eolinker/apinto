package dynamic_billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eolinker/apinto/encoder"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"
	"time"

	context_label "github.com/eolinker/apinto/utils/context-label"

	"github.com/eolinker/apinto/drivers"
	pricing_policy "github.com/eolinker/apinto/drivers/pricing-policy"
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
	extenderKeys            []string
}

// TaskInfo 描述了异步任务生成的元数据，用于二次状态查询时反查账户和资源组
type TaskInfo struct {
	ResourceID  string                      `json:"resource"`
	App         string                      `json:"app"`
	IsCharged   bool                        `json:"is_charged"`
	Variables   map[string]interface{}      `json:"variables,omitempty"`
	PricingData *pricing_policy.PricingData `json:"pricing_data,omitempty"`
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
	e.extenderKeys = cfg.ExtendKeys
	return nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
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
	app := ctx.GetLabel("application")

	// ==========================================
	// 3. 异步两阶段“查询结果 (query)”步骤的前置上下文反查还原与直接缓存拦截
	// ==========================================
	var snapshottedPricingData *pricing_policy.PricingData
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
					ctx.Response().SetStatus(http.StatusOK, "200")
					ctx.Response().SetHeader("Content-Type", "application/json")
					ctx.Response().SetBody([]byte(info.Cache.Body))
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
	// 4. 定位计费计算器
	w, has := policyManager.Get(id)
	if !has {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	calc, ok := w.Calculator().(*pricing_policy.Calculator)
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
	// 6. 组装余额扣减 Key
	// ==========================================
	balanceKey := e.balanceKeyGenerator.Key(ctx)

	// ==========================================
	// 7. 余额前置阻断校验 (Balance Pre-check)
	// ==========================================
	isPreCheckRequired := billingMode == context_label.BillingModeImmediate || billingMode == context_label.BillingModeTaskCreate
	if isPreCheckRequired && e.enableBalance && cache != nil {
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

	// ==========================================
	// 10. 组装资源定价价格 Key (优先使用固化的 PricingData，若无则从 Redis 动态读取)
	// ==========================================
	var priceData pricing_policy.PricingData
	if e.enableBalance {
		if snapshottedPricingData != nil {
			priceData = *snapshottedPricingData
		} else {
			priceKey := e.priceKeyGenerator.Key(ctx)

			strResult := cache.Get(ctx.Context(), priceKey)
			val, err := strResult.Result()
			if err != nil {
				log.Errorf("[dynamic-billing] get redis price for key %s error: %v", priceKey, err)
				return err
			}

			if err := json.Unmarshal([]byte(val), &priceData); err != nil {
				log.Errorf("[dynamic-billing] unmarshal redis pricing data error: %v, raw data: %s", err, val)
				return err
			}
			if priceData.BasicInfo != nil {
				context_label.SetPriceVersion(ctx, priceData.BasicInfo.Version)
				context_label.SetPriceVersion(ctx, priceData.BasicInfo.Rely)
			}
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
		var res *pricing_policy.CalculateResult
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
			if e.enableBalance && cache != nil && res.Sale > 0 {
				success := executeBalanceDeduction(ctx.Context(), cache, balanceKey, res.Sale, app)
				if !success {
					return p, nil
				}
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
		taskInfoKey := e.taskKeyGenerator.Key(ctx)
		info := TaskInfo{
			ResourceID: resourceID,
			App:        app,
			Variables:  calc.VariablesExtractor().ExtractAll(ctx, e.extenderKeys...),
		}

		// 获取并固化当前的定价配置数据，确保 query 阶段计费的一致性，防止中途价格配置变更
		priceKey := e.priceKeyGenerator.Key(ctx)
		if priceVal, pErr := cache.Get(ctx.Context(), priceKey).Result(); pErr == nil {
			var rData pricing_policy.PricingData
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

func (e *executor) immediateSettle(ctx http_context.IHttpContext, calc *pricing_policy.Calculator, priceData *pricing_policy.PricingData, balanceKey, resourceID, app string, cache resources.ICache) error {
	var res *pricing_policy.CalculateResult
	var err error
	res, err = calc.Calculate(ctx, e.enableBalance, priceData, e.extenderKeys...)
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

func (e *executor) settleBilling(ctx http_context.IHttpContext, calc *pricing_policy.Calculator, priceData *pricing_policy.PricingData, balanceKey, resourceID, app string, cache resources.ICache, billingMode context_label.BillingMode, isCharged bool) error {
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
			res, calcErr := calc.Calculate(ctx, e.enableBalance, priceData, e.extenderKeys...)
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
			return nil
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

// executeBalanceDeduction 在给定的 Redis 缓存上执行 Lua 脚本原子性扣减余额
func executeBalanceDeduction(ctx context.Context, cache resources.ICache, balanceKey string, cost float64, userID string) bool {
	// 余额允许为负值
	deductLua := `
		local balanceKey = KEYS[1]
		local deductAmount = tonumber(ARGV[1])
		local balanceStr = redis.call('get', balanceKey)
		local balance = 0
		if balanceStr then
			balance = tonumber(balanceStr)
		end
	
		redis.call('set', balanceKey, balance - deductAmount)
		return 1
	`
	total := cost * 1000000
	evalResCmd := cache.Run(ctx, deductLua, []string{balanceKey}, total)
	_, evalErr := evalResCmd.Result()
	if evalErr != nil {
		log.Errorf("[resource-pricing] balance deduct redis error for user %s, cost=%f: %v", userID, cost, evalErr)
		return false
	}

	// 扣款成功
	log.DebugF("[resource-pricing] balance deducted successfully for user %s, cost=%f, key=%s", userID, cost, balanceKey)
	return true
}
