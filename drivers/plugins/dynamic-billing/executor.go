package dynamic_billing

import (
	"bytes"
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
	"github.com/eolinker/apinto/drivers/credit"
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
	//enableBalance                 bool
	settingKeyGenerator           context_label.IKeyGenerator
	taskKeyGenerator              context_label.IKeyGenerator
	concurrencyKeyGenerator       context_label.IKeyGenerator
	balanceKeyGenerator           context_label.IKeyGenerator
	creditKeyGenerator            context_label.IKeyGenerator
	priceKeyGenerator             context_label.IKeyGenerator
	bindResourceGroupKeyGenerator context_label.IKeyGenerator
}

// TaskInfo 描述了异步任务生成的元数据，用于二次状态查询时反查账户和资源组
type TaskInfo struct {
	ResourceID   string                      `json:"resource"`
	App          string                      `json:"app"`
	IsCharged    bool                        `json:"is_charged"`
	PreDeductKey string                      `json:"pre_deduct_key,omitempty"`
	DeductType   int                         `json:"deduct_type,omitempty"`
	CreditID     string                      `json:"credit_id,omitempty"`
	EstCost      float64                     `json:"est_cost,omitempty"`
	CreditKey    string                      `json:"credit_key,omitempty"`
	BalanceKey   string                      `json:"balance_key,omitempty"`
	Variables    map[string]interface{}      `json:"variables,omitempty"`
	PricingData  *price_calcular.PricingData `json:"pricing_data,omitempty"`
	Cache        struct {
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
	//enableBilling = cfg.EnableBalance
	e.settingKeyGenerator = context_label.NewKeyGenerator(cfg.SettingKey)
	e.balanceKeyGenerator = context_label.NewKeyGenerator(cfg.BalanceKey)
	e.creditKeyGenerator = context_label.NewKeyGenerator(cfg.CreditKey)
	e.priceKeyGenerator = context_label.NewKeyGenerator(cfg.PriceKey)
	e.taskKeyGenerator = context_label.NewKeyGenerator(cfg.TaskKey)
	e.concurrencyKeyGenerator = context_label.NewKeyGenerator(cfg.ConcurrencyKey)
	e.bindResourceGroupKeyGenerator = context_label.NewKeyGenerator(cfg.BindResourceGroupKey)
	return nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

//
//func (e *executor) getEnableBilling() bool {
//
//}

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
	var preDeductKey string
	var preDeductType int
	var preCreditID string
	var preEstCost float64
	var restoredCreditKey string
	var restoredBalanceKey string
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

				// 还原预扣快照 key，供本次 query 阶段结算/回滚使用
				if info.PreDeductKey != "" {
					preDeductKey = info.PreDeductKey
					context_label.SetPreDeductKey(ctx, info.PreDeductKey)
				}
				if info.CreditID != "" {
					preCreditID = info.CreditID
					context_label.SetCreditID(ctx, info.CreditID)
				}
				if info.DeductType > 0 {
					preDeductType = info.DeductType
					context_label.SetDeductType(ctx, context_label.FormatDeductType(info.DeductType))
				}
				if info.EstCost > 0 {
					preEstCost = info.EstCost
					context_label.SetAmountPreDeduct(ctx, fmt.Sprintf("%f", info.EstCost))
				}
				if info.CreditKey != "" {
					restoredCreditKey = info.CreditKey
				}
				if info.BalanceKey != "" {
					restoredBalanceKey = info.BalanceKey
				}

				isCharged = info.IsCharged
				// 若已有成功计费并落盘缓存的响应，则直接写回响应体阻断拦截，防范重复调用与二次扣费
				if isCharged && info.Cache.Body != "" {
					context_label.SetAmountCost(ctx, "0")
					context_label.SetAmountSale(ctx, "0")
					context_label.SetAmountOfficial(ctx, "0")
					ctx.Response().SetStatus(http.StatusOK, "200")
					ctx.Response().SetHeader("Content-Type", "application/json")
					ctx.Response().SetBody([]byte(info.Cache.Body))
					return nil
				}
			}
		}
	}

	if resourceID == "" {
		log.Error("[dynamic-billing] resource_id not found, skipping pricing calculation.")
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	id := fmt.Sprintf("%s:%s", ctx.GetLabel("resource_type"), resourceID)

	//// 4. 定位计费计算器
	calc, ok := price_calcular.GetCalculator(id)
	if !ok || calc == nil {
		log.Errorf("[dynamic-billing] pricing-policy calculator is nil or invalid for worker %s", resourceID)
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	//context_label.SetCalculatorID(ctx, id)

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
		if cmdErr != nil {
			// Redis 出错时 fail-closed：拒绝请求，避免并发限制被绕过导致资源被刷爆。
			log.Errorf("[resource-pricing] concurrency IncrBy redis error for user %s on resource %s: %v", app, resourceID, cmdErr)
			ctx.Response().SetStatus(http.StatusServiceUnavailable, "503")
			ctx.Response().SetBody([]byte(`{"error":"concurrency check unavailable"}`))
			return nil
		}
		acquired = true
		if int(val) > concurrencyLimit {
			_ = cache.DecrBy(ctx.Context(), concurrencyKey, 1, 60*time.Second)
			log.Errorf("[resource-pricing] concurrency limit exceeded for user %s on resource %s: current=%d, limit=%d", app, resourceID, val, concurrencyLimit)
			ctx.Response().SetStatus(http.StatusTooManyRequests, "429")
			ctx.Response().SetBody([]byte(`{"error":"concurrency limit exceeded"}`))
			return nil
		}
	}

	if acquired {
		defer func() {
			_ = cache.DecrBy(ctx.Context(), concurrencyKey, 1, 60*time.Second)
		}()
	}

	settingKey := e.settingKeyGenerator.Key(ctx)
	enableBilling := false
	if customerVar != nil {
		val, has := customerVar.Get(settingKey, "enable_billing")
		if has {
			enableBilling = val == "1"
		}
	}
	// ==========================================
	// 10. 组装资源定价价格 Key (优先使用固化的 PricingData，若无则从 Redis 动态读取)
	// ==========================================
	priceData, err := e.getPricingData(ctx, cache, snapshottedPricingData, enableBilling)
	if err != nil {
		ctx.Response().SetBody([]byte(err.Error()))
		ctx.Response().SetStatus(400, "Bad Request")
		return err
	}

	// ==========================================
	// 6. 零 I/O 候选代金券规则匹配与 Key 组装
	// ==========================================
	resourceType := ctx.GetLabel("resource_type")
	provider := ctx.GetLabel("provider")
	if provider == "" {
		provider = ctx.GetLabel("service")
	}

	candidateRules := credit.MatchCandidateRules(resourceType, provider, resourceID, time.Now().Unix())
	candidateCreditIDs := make([]string, 0, len(candidateRules))
	for _, r := range candidateRules {
		candidateCreditIDs = append(candidateCreditIDs, r.ID)
	}

	balanceKey := e.balanceKeyGenerator.Key(ctx)
	if restoredBalanceKey != "" {
		balanceKey = restoredBalanceKey
	}
	creditKey := e.creditKeyGenerator.Key(ctx)
	if restoredCreditKey != "" {
		creditKey = restoredCreditKey
	}

	// ==========================================
	// 7. 余额前置阻断校验 (Balance Pre-check)
	//    若定价数据中存在任一"免费策略"，跳过余额拦截；
	//    若持有候选代金券，则放行至阶段一原子预扣脚本进行扣减判定；
	//    否则按余额<=0 拦截。
	// ==========================================
	isPreCheckRequired := billingMode == context_label.BillingModeImmediate || billingMode == context_label.BillingModeTaskCreate
	if isPreCheckRequired && enableBilling && cache != nil && !hasFreePricePlan(priceData) && len(candidateCreditIDs) == 0 {
		balanceStr, err := cache.Get(ctx.Context(), balanceKey).Result()
		if err == nil {
			if balanceInt, parseErr := strconv.ParseInt(balanceStr, 10, 64); parseErr == nil {
				if balanceInt <= 0 {
					balance := float64(balanceInt) / 1000000.0
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

	// ==========================================
	// 7.5 阶段一 原子预扣 Pre-deduct（优先抵扣最先失效代金券，不足原子回退现金余额）
	// ==========================================
	if enableBilling && cache != nil && isPreCheckRequired && !hasFreePricePlan(priceData) {
		amt, rid, pdErr := calc.PreDeduct(ctx, priceData)
		if pdErr != nil {
			log.Errorf("[dynamic-billing] pre-deduct calculate error for resource %s: %v", resourceID, pdErr)
		} else if amt > 0 {
			preDeductKey = preDeductKeyGenerator.Key(ctx, func(c eocontext.EoContext, label string) string {
				if label == "request_id" {
					return c.RequestId()
				}
				return ""
			})
			ttl := preDeductTTL
			if billingMode == context_label.BillingModeTaskCreate {
				ttl = 24 * time.Hour
			}
			ok, dType, cID := executePreDeductTwoPhase(ctx.Context(), cache, creditKey, balanceKey, preDeductKey, candidateCreditIDs, amt, ttl, app)
			if !ok {
				log.Errorf("[dynamic-billing] pre-deduct failed for user %s, amount=%f", app, amt)
				ctx.Response().SetStatus(http.StatusPaymentRequired, "402")
				ctx.Response().SetBody([]byte(`{"error":"insufficient balance"}`))
				return errors.New(`{"error":"insufficient balance"}`)
			}
			preDeductType = dType
			preCreditID = cID
			preEstCost = amt
			context_label.SetAmountPreDeduct(ctx, fmt.Sprintf("%f", amt))
			context_label.SetPreDeductKey(ctx, preDeductKey)
			if dType == 1 && cID != "" {
				context_label.SetCreditID(ctx, cID)
				context_label.SetDeductType(ctx, context_label.DeductTypeCredit)
			} else if dType == 3 && cID != "" {
				context_label.SetCreditID(ctx, cID)
				context_label.SetDeductType(ctx, context_label.DeductTypeCombined)
			} else {
				context_label.SetDeductType(ctx, context_label.DeductTypeBalance)
			}
			log.DebugF("[dynamic-billing] pre-deducted %f (type=%d, credit=%s) for user %s on resource %s (rule=%s, key=%s)",
				amt, dType, cID, app, resourceID, rid, preDeductKey)
		}
	}

	// 阶段二 响应结算回调（仅对 Immediate 模式注册）
	if billingMode == context_label.BillingModeImmediate && ctx.GetLabel("resource_type") == "api" {
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			if ctx.Response().StatusCode() >= 500 || ctx.Response().ResponseError() != nil {
				log.Warnf("[dynamic-billing] upstream 5xx/error for user %s on resource %s, triggering failure recovery", app, resourceID)
				triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
				return
			}
			res, err := calc.Calculate(ctx, enableBilling, priceData)
			if err != nil {
				log.Errorf("[dynamic-billing] calculate error: %v", err)
				triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
				return
			}
			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if res.Marked > 0 {
				context_label.SetAmountMarked(ctx, fmt.Sprintf("%f", res.Marked))
			}
			if enableBilling && cache != nil {
				settlePreDeductTwoPhase(ctx, cache, creditKey, balanceKey, preDeductKey, candidateCreditIDs, res.Sale, app)
			}
		})
	} else if billingMode == context_label.BillingModeImmediate && ctx.GetLabel("resource_type") == "ai" {
		var res *price_calcular.CalculateResult
		var streamBuffer []byte
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			if ctx.Response().StatusCode() >= 500 || ctx.Response().ResponseError() != nil {
				log.Warnf("[dynamic-billing] upstream 5xx/error for user %s on ai resource %s, triggering failure recovery", app, resourceID)
				triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
				return
			}
			if len(streamBuffer) > 0 {
				line := strings.TrimSpace(string(streamBuffer))
				if line != "" {
					tmp, err := calc.CalculateFromChunk(ctx, enableBilling, []byte(line), priceData)
					if err == nil && tmp != nil {
						res = tmp
					}
				}
				streamBuffer = nil
			}
			if res != nil && res.Sale > 0 {
				context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
				context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
				context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
				if res.Marked > 0 {
					context_label.SetAmountMarked(ctx, fmt.Sprintf("%f", res.Marked))
				}
				if enableBilling && cache != nil {
					settlePreDeductTwoPhase(ctx, cache, creditKey, balanceKey, preDeductKey, candidateCreditIDs, res.Sale, app)
				}
			} else {
				// 未完成实际结算：全额冲正预扣
				triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
			}
		})
		// 只有文本模型需要异步流式处理
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
			// 拼接到请求级流式缓冲区
			streamBuffer = append(streamBuffer, body...)

			// 查找最后一个换行符
			lastNL := bytes.LastIndexByte(streamBuffer, '\n')
			if lastNL == -1 {
				return p, nil
			}

			// 截取所有完整行，残余未结束的部分留待下一个 Chunk 拼接
			completeData := streamBuffer[:lastNL]
			streamBuffer = append([]byte(nil), streamBuffer[lastNL+1:]...)

			// 遍历已接收完整的各行
			lines := strings.Split(string(completeData), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				tmp, err := calc.CalculateFromChunk(ctx, enableBilling, []byte(line), priceData)
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
			if res.Marked > 0 {
				context_label.SetAmountMarked(ctx, fmt.Sprintf("%f", res.Marked))
			}

			return p, nil
		})
	}
	_ = preDeductKey

	// ==========================================
	// 8. 转发下游链路 (Do Chain Forwarding)
	// ==========================================
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			log.Errorf("[dynamic-billing] downstream chain forwarding error: %v", err)
			triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
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
			ResourceID:   resourceID,
			App:          app,
			PreDeductKey: preDeductKey,
			DeductType:   preDeductType,
			CreditID:     preCreditID,
			EstCost:      preEstCost,
			CreditKey:    creditKey,
			BalanceKey:   balanceKey,
			Variables:    extractor.ExtractAll(ctx),
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

	// 同步结算非流式响应（Immediate 非流式请求 & TaskQuery 阶段）
	return e.settleBilling(ctx, calc, priceData, creditKey, balanceKey, candidateCreditIDs, preDeductKey, preDeductType, preCreditID, preEstCost, resourceID, app, cache, billingMode, isCharged, enableBilling)
}

func (e *executor) getPricingData(ctx http_context.IHttpContext, cache resources.ICache, snapPricingData *price_calcular.PricingData, enableBilling bool) (*price_calcular.PricingData, error) {
	var priceData *price_calcular.PricingData
	if enableBilling {
		if snapPricingData != nil {
			priceData = snapPricingData
		} else {
			isUser := context_label.IsUserConsumer(ctx)
			if !isUser {
				priceKey := e.priceKeyGenerator.Key(ctx)

				strResult := cache.Get(ctx.Context(), priceKey)
				val, err := strResult.Result()
				if err != nil {
					log.Errorf("[dynamic-billing] get redis price for key %s error: %v", priceKey, err)
					return nil, fmt.Errorf("{\"error\":\"pricing(%s) data not found\"}", priceKey)
				}
				var pd price_calcular.PricingData
				if err := json.Unmarshal([]byte(val), &pd); err != nil {
					log.Errorf("[dynamic-billing] unmarshal redis pricing data error: %v, raw data: %s", err, val)
					return nil, fmt.Errorf("{\"error\":\"invalid pricing(%s) data\"}", priceKey)
				}
				priceData = &pd
			} else {
				if customerVar == nil {
					return nil, fmt.Errorf("{\"error\":\"no resource group found\"}")
				}
				key := e.bindResourceGroupKeyGenerator.Key(ctx)
				rgs, ok := customerVar.GetAll(key)
				if !ok {
					return nil, fmt.Errorf("{\"error\":\"no resource group found\"}")
				}
				pds := make([]*price_calcular.PricingData, 0, len(rgs))
				for key := range rgs {
					priceKey := e.priceKeyGenerator.Key(ctx, func(ctx eocontext.EoContext, label string) string {
						if label == "application" {
							return key
						}
						return ""
					})
					strResult := cache.Get(ctx.Context(), priceKey)
					val, err := strResult.Result()
					if err != nil {
						continue
					}
					var pd price_calcular.PricingData
					if err = json.Unmarshal([]byte(val), &pd); err != nil {
						continue
					}
					pds = append(pds, &pd)
				}
				priceData = price_calcular.MinSalePricingData(pds)
			}
		}
	}
	if priceData != nil {
		if priceData.BasicInfo != nil {
			context_label.SetPriceVersion(ctx, priceData.BasicInfo.Version)
			context_label.SetPriceRelyVersion(ctx, priceData.BasicInfo.Rely)
		}
	} else if enableBilling {
		return nil, fmt.Errorf("{\"error\":\"pricing data not found\"}")
	}
	return priceData, nil
}

func (e *executor) immediateSettle(ctx http_context.IHttpContext, calc price_calcular.ICalculator, priceData *price_calcular.PricingData, creditKey, balanceKey string, candidateCreditIDs []string, preDeductKey string, preDeductType int, preCreditID string, preEstCost float64, resourceID, app string, cache resources.ICache, enableBilling bool) error {
	var res *price_calcular.CalculateResult
	var err error
	res, err = calc.Calculate(ctx, enableBilling, priceData)
	if err != nil {
		log.Errorf("[dynamic-billing] calculate failed for resource %s: %v", resourceID, err)
		triggerRecovery(ctx, cache, creditKey, balanceKey, preDeductKey, preDeductType, preCreditID, preEstCost, app)
		return err
	}

	context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
	context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
	context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
	if res.Marked > 0 {
		context_label.SetAmountMarked(ctx, fmt.Sprintf("%f", res.Marked))
	}

	if enableBilling && cache != nil {
		settlePreDeductTwoPhase(ctx, cache, creditKey, balanceKey, preDeductKey, candidateCreditIDs, res.Sale, app)
	}
	return nil
}

func (e *executor) settleBilling(ctx http_context.IHttpContext, calc price_calcular.ICalculator, priceData *price_calcular.PricingData, creditKey, balanceKey string, candidateCreditIDs []string, preDeductKey string, preDeductType int, preCreditID string, preEstCost float64, resourceID, app string, cache resources.ICache, billingMode context_label.BillingMode, isCharged bool, enableBilling bool) error {
	switch billingMode {
	case context_label.BillingModeImmediate:
		return e.immediateSettle(ctx, calc, priceData, creditKey, balanceKey, candidateCreditIDs, preDeductKey, preDeductType, preCreditID, preEstCost, resourceID, app, cache, enableBilling)
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
		// TaskQuery 阶段结算/回滚 都需要 TaskCreate 阶段落下的预扣快照 key。
		// 它已在 DoHttpFilter 反查 TaskInfo 时被写回上下文，这里直接读取即可。
		snapshotKey := context_label.GetPreDeductKey(ctx)
		if snapshotKey == "" {
			snapshotKey = preDeductKey
		}
		switch taskStatus {
		case context_label.TaskStatusRunning:
		case context_label.TaskStatusSuccess:
			isCharged = true
			res, calcErr := calc.Calculate(ctx, enableBilling, priceData)
			if calcErr != nil {
				log.Errorf("[dynamic-billing] calculate failed for resource %s: %v", resourceID, calcErr)
				ctx.SetLabel("pricing_status", "failed")
				ctx.SetLabel("pricing_error", calcErr.Error())
				triggerRecovery(ctx, cache, creditKey, balanceKey, snapshotKey, preDeductType, preCreditID, preEstCost, app)
				return calcErr
			}

			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if res.Marked > 0 {
				context_label.SetAmountMarked(ctx, fmt.Sprintf("%f", res.Marked))
			}
			if enableBilling && cache != nil {
				// 任务执行成功：以预扣快照做多退少补两阶段跨账本结算
				settlePreDeductTwoPhase(ctx, cache, creditKey, balanceKey, snapshotKey, candidateCreditIDs, res.Sale, app)
			}
		case context_label.TaskStatusFailed:
			isCharged = true
			// 任务执行失败：整额冲正退回 TaskCreate 阶段的预扣金额。
			if enableBilling && cache != nil {
				triggerRecovery(ctx, cache, creditKey, balanceKey, snapshotKey, preDeductType, preCreditID, preEstCost, app)
			}
		}
		taskKey := e.taskKeyGenerator.Key(ctx)
		taskInfo := TaskInfo{
			ResourceID: resourceID,
			App:        app,
			IsCharged:  isCharged,
			// 结算/回滚已完成后，预扣快照已被 DEL；清空 PreDeductKey，
			// 避免后续幂等重入的 query 请求误用一个已失效的 key。
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
	// 余额与扣减金额统一以「元 × 1000000」的整数存储（与前置校验的 /1000000.0 保持一致），
	// 保留 6 位小数精度。先四舍五入消除浮点误差后转为整数，避免向 Redis 传入浮点。
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

// settlePreDeduct 依据预扣快照与实际结算金额进行"多退少补"。
//
//   - actualSale >  preDeducted：追加扣除差额（少补）
//   - actualSale <  preDeducted：把差额退还给用户余额（多退）
//   - actualSale == preDeducted：无操作
//   - 若不存在预扣快照（preDeductKey 为空或已被结算过），则退化为按实际金额单独扣除
//
// 高并发场景下，本函数以单条 Lua 脚本原子完成 "GET pre → DEL pre → DECRBY balance diff"，
// DEL 保证了脚本天然幂等 —— 同一请求即使因回调重入被多次调用，账面变化只会发生一次。
//
// 关于"少补"时的余额不足：由于服务已经产出，此处不再拦截扣款，允许余额透支为负，
// 由上层欠费流程处理；同时打 warn 日志方便对账。"多退"时差额为负 (DECRBY 负值 = INCRBY)，
// 不受余额限制，天然安全。
func settlePreDeduct(ctx http_context.IHttpContext, cache resources.ICache, balanceKey, preDeductKey string, actualSale float64, userID string) {
	if cache == nil {
		log.Errorf("[resource-pricing] settlePreDeduct failed: cache is nil for user %s", userID)
		return
	}
	actual := int64(math.Round(actualSale * 1000000))

	if preDeductKey == "" {
		// 没有预扣，退化为普通扣款
		if actual > 0 {
			executeBalanceDeduction(ctx.Context(), cache, balanceKey, actualSale, userID)
		}
		return
	}

	// Lua 脚本原子完成：
	//   pre = GET preKey (缺省 0)
	//   若 pre == 0：说明已被结算/回滚过，直接返回幂等结果
	//   否则 DEL preKey；diff = actual - pre；DECRBY balance diff
	// 返回 {applied_diff, new_balance, pre}
	settleLua := `
		local balanceKey = KEYS[1]
		local preKey = KEYS[2]
		local actual = tonumber(ARGV[1])
		local pre = tonumber(redis.call('get', preKey) or "0")
		if pre == 0 then
			return {0, tonumber(redis.call('get', balanceKey) or "0"), 0}
		end
		redis.call('del', preKey)
		local diff = actual - pre
		if diff == 0 then
			return {0, tonumber(redis.call('get', balanceKey) or "0"), pre}
		end
		local nb = redis.call('decrby', balanceKey, diff)
		return {diff, nb, pre}
	`
	res := cache.Run(ctx.Context(), settleLua, []string{balanceKey, preDeductKey}, actual)
	raw, err := res.Result()
	if err != nil {
		log.Errorf("[dynamic-billing] settle pre-deduct redis error for user %s, actual=%f, key=%s: %v",
			userID, actualSale, preDeductKey, err)
		return
	}
	// 少补时如果导致余额为负，打 warn 日志便于人工/欠费流程处理
	if arr, ok := raw.([]interface{}); ok && len(arr) >= 2 {
		if nb, ok2 := arr[1].(int64); ok2 && nb < 0 {
			log.Warnf("[dynamic-billing] balance turned negative after settle for user %s, key=%s, balance=%d(×1e6)",
				userID, preDeductKey, nb)
		}
	}
	log.DebugF("[dynamic-billing] settled pre-deduct for user %s, actual=%f, key=%s", userID, actualSale, preDeductKey)
}

// refundPreDeduct 全额退回此前的预扣金额。计算失败或未能产生实际结算时调用。
//
// 使用一条 Lua 脚本原子完成 "GET pre → DEL pre → INCRBY balance pre"，
// 幂等：key 不存在时脚本直接返回，不会重复退款。退款是加钱操作，不受当前余额限制，
// 无需额外校验。
func refundPreDeduct(ctx http_context.IHttpContext, cache resources.ICache, balanceKey, preDeductKey, userID string) {
	if cache == nil || preDeductKey == "" {
		return
	}
	refundLua := `
		local balanceKey = KEYS[1]
		local preKey = KEYS[2]
		local pre = tonumber(redis.call('get', preKey) or "0")
		if pre == 0 then
			return {0, tonumber(redis.call('get', balanceKey) or "0")}
		end
		redis.call('del', preKey)
		local nb = redis.call('incrby', balanceKey, pre)
		return {pre, nb}
	`
	res := cache.Run(ctx.Context(), refundLua, []string{balanceKey, preDeductKey})
	if _, err := res.Result(); err != nil {
		log.Errorf("[dynamic-billing] refund pre-deduct redis error for user %s, key=%s: %v", userID, preDeductKey, err)
		return
	}
	log.DebugF("[dynamic-billing] refunded pre-deduct for user %s, key=%s", userID, preDeductKey)
}

// executePreDeduct 以单条 Lua 脚本原子完成 "余额校验 + 余额 DECRBY N + SETEX preKey N ttl"。
//
// 高并发要点：
//   - 全部动作在 Redis 单线程内一次执行完，天然是原子的：
//     1) 先 GET 当前余额；
//     2) 若余额 < amount，则直接返回失败，不扣款、不落快照；
//     3) 否则 DECRBY balance amount，并 SET preKey amount EX ttl。
//     这从根本上杜绝了 "两个并发请求先后通过前置 balance>0 校验然后各自 DECRBY 造成
//     余额穿透至负数" 的竞态。
//   - preKey 通常包含 request_id，不同请求间互不冲突；
//     SETEX 保证快照最长存活 preDeductTTL，异常退出的孤儿快照能自然过期，
//     不会永久占用用户余额。
//
// 返回 true 表示成功扣款且已落快照；返回 false 表示 Redis 出错或余额不足，
// 调用方应据此拦截并回复 402。
//
// ttl 由调用方传入：同步 immediate 场景用 preDeductTTL（分钟级），
// 异步 TaskCreate 场景需与 TaskInfo 缓存同寿（24h），
// 否则 query 阶段到来时预扣快照可能已经过期导致无法结算/回滚。
func executePreDeduct(ctx context.Context, cache resources.ICache, balanceKey, preDeductKey string, amount float64, ttl time.Duration, userID string) bool {
	preDeductLua := `
		local balanceKey = KEYS[1]
		local preKey = KEYS[2]
		local amount = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])
		local balance = tonumber(redis.call('get', balanceKey) or "0")
		if balance < amount then
			return {0, balance}
		end
		local nb = redis.call('decrby', balanceKey, amount)
		redis.call('set', preKey, amount, 'EX', ttl)
		return {1, nb}
	`
	amt := int64(math.Round(amount * 1000000))
	if ttl <= 0 {
		ttl = preDeductTTL
	}
	res := cache.Run(ctx, preDeductLua, []string{balanceKey, preDeductKey}, amt, int64(ttl/time.Second))
	raw, err := res.Result()
	if err != nil {
		log.Errorf("[dynamic-billing] pre-deduct redis error for user %s, amount=%f, key=%s: %v",
			userID, amount, preDeductKey, err)
		return false
	}
	// 解析返回值 {ok, balance}：Lua 脚本明确返回 {1, nb}（成功）或 {0, balance}（余额不足）。
	// 返回值类型不符合预期时视为失败（fail-closed），避免没扣钱却放行请求导致账面错乱。
	arr, ok := raw.([]interface{})
	if !ok || len(arr) < 2 {
		log.Errorf("[dynamic-billing] pre-deduct unexpected redis return for user %s, amount=%f, key=%s, raw=%v",
			userID, amount, preDeductKey, raw)
		return false
	}
	okFlag, _ := arr[0].(int64)
	if okFlag == 0 {
		// 余额不足
		log.Errorf("[dynamic-billing] pre-deduct rejected: insufficient balance for user %s, amount=%f, key=%s",
			userID, amount, preDeductKey)
		return false
	}
	return true
}

// preDeductTTL 预扣快照的最大存活时间。请求处理超时或进程崩溃时，孤儿快照会
// 在此时限内被 Redis 自动清理，避免占用用户余额。
const preDeductTTL = 5 * time.Minute

var (
	preDeductKeyGenerator = context_label.NewKeyGenerator("{{product}:balance}:pre-deduct:{request_id}")
)

// =========================================================================
// 7.2 阶段一 原子预扣 Pre-deduct（优先抵扣最先失效代金券，不足时代金券扣至0，差额扣现金）
// =========================================================================

// executePreDeductTwoPhase 执行两阶段原子预扣：
//  1. 遍历 ARGV[3...N] 中的候选券 ID，寻找第一张有余额 (cBal > 0) 的候选券；
//     a. 若券余额充足 (cBal >= amt)，全额扣减代金券，记录快照 {deduct_type: 1, credit_id: cid, amount: amt, credit_amount: amt, balance_amount: 0}
//     b. 若券余额不足 (0 < cBal < amt)，将券余额预扣至 0，不足差额检查现金余额 (bBal >= remainAmt) 并扣减现金，
//     记录快照 {deduct_type: 3, credit_id: cid, amount: amt, credit_amount: cBal, balance_amount: remainAmt}
//     若现金不足以支付差额，则原子阻断返回 0。
//  2. 若无候选券或所有候选券余额均为 0，检查现金余额 balanceKey >= amt，若充足则扣减现金，
//     记录快照 {deduct_type: 2, credit_id: ”, amount: amt, credit_amount: 0, balance_amount: amt}
//  3. 均不足则拦截返回 0，由上层阻断响应 402 Payment Required。
func executePreDeductTwoPhase(ctx context.Context, cache resources.ICache, creditKey, balanceKey, preDeductKey string, candidateCreditIDs []string, amount float64, ttl time.Duration, userID string) (bool, int, string) {
	if cache == nil {
		return false, 0, ""
	}
	amt := int64(math.Round(amount * 1000000))
	if amt <= 0 {
		return true, 0, ""
	}
	if ttl <= 0 {
		ttl = preDeductTTL
	}

	preDeductLua := `
		local creditKey = KEYS[1]
		local balanceKey = KEYS[2]
		local preKey = KEYS[3]
		local amt = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])

		-- 1. 寻找第一张有余额 (cBal > 0) 的候选代金券 (ARGV[3...N])
		local targetCid = ""
		local targetCBal = 0
		for i = 3, #ARGV do
			local cid = ARGV[i]
			local cBal = tonumber(redis.call('hget', creditKey, cid) or "0")
			if cBal > 0 then
				targetCid = cid
				targetCBal = cBal
				break
			end
		end

		if targetCid ~= "" then
			if targetCBal >= amt then
				-- 1.1 代金券余额充足：全额扣减代金券
				local nb = redis.call('hincrby', creditKey, targetCid, -amt)
				redis.call('hmset', preKey, 
					'deduct_type', 1, 
					'credit_id', targetCid, 
					'amount', amt, 
					'credit_amount', amt, 
					'balance_amount', 0
				)
				redis.call('expire', preKey, ttl)
				return {1, 1, targetCid, nb, 0}
			else
				-- 1.2 代金券余额不足预扣金额：代金券扣到 0，不足部分检查并扣减现金余额
				local remainAmt = amt - targetCBal
				local bBal = tonumber(redis.call('get', balanceKey) or "-1")
				if bBal >= remainAmt then
					redis.call('hincrby', creditKey, targetCid, -targetCBal)
					local nb = redis.call('decrby', balanceKey, remainAmt)
					redis.call('hmset', preKey, 
						'deduct_type', 3, 
						'credit_id', targetCid, 
						'amount', amt, 
						'credit_amount', targetCBal, 
						'balance_amount', remainAmt
					)
					redis.call('expire', preKey, ttl)
					return {1, 3, targetCid, 0, nb}
				else
					-- 代金券+现金余额总额不足预扣金额：原子阻断
					return {0, 0, '', targetCBal, bBal}
				end
			end
		else
			-- 2. 无候选代金券或所有代金券余额为 0：全额检查并扣减现金余额
			local bBal = tonumber(redis.call('get', balanceKey) or "-1")
			if bBal >= amt then
				local nb = redis.call('decrby', balanceKey, amt)
				redis.call('hmset', preKey, 
					'deduct_type', 2, 
					'credit_id', '', 
					'amount', amt, 
					'credit_amount', 0, 
					'balance_amount', amt
				)
				redis.call('expire', preKey, ttl)
				return {1, 2, '', 0, nb}
			else
				-- 现金余额不足，阻断
				return {0, 0, '', 0, bBal}
			end
		end
	`

	args := make([]interface{}, 0, 2+len(candidateCreditIDs))
	args = append(args, amt, int64(ttl/time.Second))
	for _, cid := range candidateCreditIDs {
		args = append(args, cid)
	}

	res := cache.Run(ctx, preDeductLua, []string{creditKey, balanceKey, preDeductKey}, args...)
	raw, err := res.Result()
	if err != nil {
		log.Errorf("[dynamic-billing] pre-deduct two-phase redis error for user %s, amount=%f, key=%s: %v",
			userID, amount, preDeductKey, err)
		return false, 0, ""
	}

	arr, ok := raw.([]interface{})
	if !ok || len(arr) < 3 {
		log.Errorf("[dynamic-billing] pre-deduct two-phase unexpected return for user %s: %v", userID, raw)
		return false, 0, ""
	}

	okFlag, _ := arr[0].(int64)
	if okFlag == 0 {
		log.Errorf("[dynamic-billing] pre-deduct rejected: insufficient credit and balance for user %s, amount=%f", userID, amount)
		return false, 0, ""
	}

	deductType, _ := arr[1].(int64)
	creditID, _ := arr[2].(string)
	return true, int(deductType), creditID
}

// =========================================================================
// 7.3 阶段二 响应结算与跨账本抵扣 Post-settlement (多退少补与代金券转扣现金)
// =========================================================================

// settlePreDeductTwoPhase 依据真实消费金额结算预扣快照：
// 1. 若无预扣快照（已结算/回滚过或未预扣），按真实金额扣减：优先扣代金券至0，不足扣现金
// 2. 若快照 deduct_type == 1 (代金券预扣)：
//   - sale <= preAmt (多退)：差额原路退还代金券 (hincrby creditKey credit_id (preAmt - sale))
//   - sale > preAmt (少补)：
//     当前代金券余额充足则追加扣减代金券；
//     代金券余额不足，将代金券余额清零，剩余未扣差额全额转扣现金余额 balanceKey（允许负数透支）
//
// 3. 若快照 deduct_type == 2 (现金预扣)：
//   - 现金余额多退少补：decrby balanceKey (sale - preAmt)
//
// 4. 若快照 deduct_type == 3 (跨账本预扣：券 preCreditAmt + 现金 preBalanceAmt)：
//   - 多退 (diff < 0)：优先退现金，现金退完若仍有多余再退还代金券
//   - 少补 (diff > 0)：优先追加代金券，代金券不足则追加扣减现金（允许透支为负）
//
// 5. 原子 DEL preKey 保障绝对幂等
func settlePreDeductTwoPhase(ctx http_context.IHttpContext, cache resources.ICache, creditKey, balanceKey, preDeductKey string, candidateCreditIDs []string, actualSale float64, userID string) {
	if cache == nil {
		return
	}
	actual := int64(math.Round(actualSale * 1000000))

	if preDeductKey == "" {
		// 无预扣快照，退化为直接抵扣（同样遵循券扣到0不足扣现金）
		if actual <= 0 {
			return
		}
		directDeductLua := `
			local creditKey = KEYS[1]
			local balanceKey = KEYS[2]
			local actual = tonumber(ARGV[1])

			local targetCid = ""
			local targetCBal = 0
			for i = 2, #ARGV do
				local candCid = ARGV[i]
				local cBal = tonumber(redis.call('hget', creditKey, candCid) or "0")
				if cBal > 0 then
					targetCid = candCid
					targetCBal = cBal
					break
				end
			end

			if targetCid ~= "" then
				if targetCBal >= actual then
					local nb = redis.call('hincrby', creditKey, targetCid, -actual)
					return {1, 1, targetCid, actual, 0}
				else
					local remain = actual - targetCBal
					redis.call('hincrby', creditKey, targetCid, -targetCBal)
					local nb = redis.call('decrby', balanceKey, remain)
					return {1, 3, targetCid, targetCBal, remain, nb}
				end
			else
				local nb = redis.call('decrby', balanceKey, actual)
				return {1, 2, '', 0, actual, nb}
			end
		`
		args := make([]interface{}, 0, 1+len(candidateCreditIDs))
		args = append(args, actual)
		for _, cid := range candidateCreditIDs {
			args = append(args, cid)
		}
		res := cache.Run(ctx.Context(), directDeductLua, []string{creditKey, balanceKey}, args...)
		raw, err := res.Result()
		if err != nil {
			log.Errorf("[dynamic-billing] direct deduct redis error for user %s, actual=%f: %v", userID, actualSale, err)
			return
		}
		if arr, ok := raw.([]interface{}); ok && len(arr) >= 4 {
			settleDType, _ := arr[1].(int64)
			usedCID, _ := arr[2].(string)
			switch settleDType {
			case 1:
				context_label.SetDeductType(ctx, context_label.DeductTypeCredit)
				if usedCID != "" {
					context_label.SetCreditID(ctx, usedCID)
				}
				context_label.SetCreditApplied(ctx, fmt.Sprintf("%f", actualSale))
				context_label.SetBalanceApplied(ctx, "0")
			case 2:
				context_label.SetDeductType(ctx, context_label.DeductTypeBalance)
				context_label.SetCreditApplied(ctx, "0")
				context_label.SetBalanceApplied(ctx, fmt.Sprintf("%f", actualSale))
			case 3:
				context_label.SetDeductType(ctx, context_label.DeductTypeCombined)
				if usedCID != "" {
					context_label.SetCreditID(ctx, usedCID)
				}
				creditAmt, _ := arr[3].(int64)
				balanceAmt, _ := arr[4].(int64)
				context_label.SetCreditApplied(ctx, fmt.Sprintf("%f", float64(creditAmt)/1000000.0))
				context_label.SetBalanceApplied(ctx, fmt.Sprintf("%f", float64(balanceAmt)/1000000.0))
			}
		}
		return
	}

	settleLua := `
		local creditKey = KEYS[1]
		local balanceKey = KEYS[2]
		local preKey = KEYS[3]
		local actual = tonumber(ARGV[1])

		-- 读取并原子删除预扣快照保证幂等
		local snap = redis.call('hmget', preKey, 'deduct_type', 'credit_id', 'amount', 'credit_amount', 'balance_amount')
		redis.call('del', preKey)

		local dType = tonumber(snap[1] or "0")
		local cid = snap[2] or ""
		local preAmt = tonumber(snap[3] or "0")
		local preCreditAmt = tonumber(snap[4] or "0")
		local preBalanceAmt = tonumber(snap[5] or "0")

		if preCreditAmt == 0 and preBalanceAmt == 0 then
			if dType == 1 then
				preCreditAmt = preAmt
			elseif dType == 2 then
				preBalanceAmt = preAmt
			end
		end

		if dType == 0 or preAmt == 0 then
			-- 快照不存在或已被结算过，幂等直接返回
			return {1, 0, '', 0, 0}
		end

		local diff = actual - preAmt

		if dType == 1 then
			-- 原预扣为纯代金券
			if diff <= 0 then
				-- 多退：退还代金券
				local refund = -diff
				if refund > 0 then
					redis.call('hincrby', creditKey, cid, refund)
				end
				return {1, 1, cid, actual, 0}
			else
				-- 少补：diff > 0，优先追加代金券
				local cBal = tonumber(redis.call('hget', creditKey, cid) or "0")
				if cBal >= diff then
					redis.call('hincrby', creditKey, cid, -diff)
					return {1, 1, cid, actual, 0}
				else
					-- 代金券不足：代金券清零，剩余差额转扣主账户现金余额（允许透支为负）
					local deductCredit = cBal
					if deductCredit > 0 then
						redis.call('hincrby', creditKey, cid, -deductCredit)
					else
						deductCredit = 0
					end
					local remainDiff = diff - deductCredit
					local nb = redis.call('decrby', balanceKey, remainDiff)
					return {1, 3, cid, preAmt + deductCredit, remainDiff, nb}
				end
			end
		elseif dType == 3 then
			-- 原预扣为跨账本 (券 preCreditAmt + 现金 preBalanceAmt)
			if diff < 0 then
				-- 多退：优先退还现金，现金退完仍有多余退代金券
				local refund = -diff
				if refund <= preBalanceAmt then
					redis.call('incrby', balanceKey, refund)
					local actualBalance = preBalanceAmt - refund
					return {1, 3, cid, preCreditAmt, actualBalance}
				else
					redis.call('incrby', balanceKey, preBalanceAmt)
					local remainRefund = refund - preBalanceAmt
					redis.call('hincrby', creditKey, cid, remainRefund)
					return {1, 1, cid, actual, 0}
				end
			elseif diff > 0 then
				-- 少补：优先尝试追加代金券
				local cBal = tonumber(redis.call('hget', creditKey, cid) or "0")
				if cBal >= diff then
					redis.call('hincrby', creditKey, cid, -diff)
					return {1, 3, cid, preCreditAmt + diff, preBalanceAmt}
				else
					local deductCredit = cBal
					if deductCredit > 0 then
						redis.call('hincrby', creditKey, cid, -deductCredit)
					else
						deductCredit = 0
					end
					local remainDiff = diff - deductCredit
					local nb = redis.call('decrby', balanceKey, remainDiff)
					return {1, 3, cid, preCreditAmt + deductCredit, preBalanceAmt + remainDiff, nb}
				end
			else
				-- 等额
				return {1, 3, cid, preCreditAmt, preBalanceAmt}
			end
		else
			-- 原预扣为纯现金 (dType == 2)
			if diff ~= 0 then
				local nb = redis.call('decrby', balanceKey, diff)
				return {1, 2, '', 0, actual, nb}
			end
			return {1, 2, '', 0, actual}
		end
	`

	args := make([]interface{}, 0, 1+len(candidateCreditIDs))
	args = append(args, actual)
	for _, cid := range candidateCreditIDs {
		args = append(args, cid)
	}

	res := cache.Run(ctx.Context(), settleLua, []string{creditKey, balanceKey, preDeductKey}, args...)
	raw, err := res.Result()
	if err != nil {
		log.Errorf("[dynamic-billing] settle pre-deduct two-phase redis error for user %s, actual=%f, key=%s: %v",
			userID, actualSale, preDeductKey, err)
		return
	}

	// 记录计费结算上下文标签
	if arr, ok := raw.([]interface{}); ok && len(arr) >= 4 {
		settleDType, _ := arr[1].(int64)
		usedCID, _ := arr[2].(string)
		switch settleDType {
		case 1:
			// 全额代金券抵扣
			context_label.SetDeductType(ctx, context_label.DeductTypeCredit)
			if usedCID != "" {
				context_label.SetCreditID(ctx, usedCID)
			}
			context_label.SetCreditApplied(ctx, fmt.Sprintf("%f", actualSale))
			context_label.SetBalanceApplied(ctx, "0")
		case 2:
			// 全额主账户现金抵扣
			context_label.SetDeductType(ctx, context_label.DeductTypeBalance)
			context_label.SetCreditApplied(ctx, "0")
			context_label.SetBalanceApplied(ctx, fmt.Sprintf("%f", actualSale))
		case 3:
			// 跨账本抵扣：代金券不足清零 + 主账户现金转扣
			context_label.SetDeductType(ctx, context_label.DeductTypeCombined)
			if usedCID != "" {
				context_label.SetCreditID(ctx, usedCID)
			}
			creditAmt, _ := arr[3].(int64)
			balanceAmt, _ := arr[4].(int64)
			context_label.SetCreditApplied(ctx, fmt.Sprintf("%f", float64(creditAmt)/1000000.0))
			context_label.SetBalanceApplied(ctx, fmt.Sprintf("%f", float64(balanceAmt)/1000000.0))
			if len(arr) >= 6 {
				if nb, ok2 := arr[5].(int64); ok2 && nb < 0 {
					log.Warnf("[dynamic-billing] balance turned negative after cross-ledger settlement for user %s: balance=%d(×1e6)",
						userID, nb)
				}
			}
		}
	}
	log.DebugF("[dynamic-billing] settled pre-deduct two-phase for user %s, actual=%f, key=%s", userID, actualSale, preDeductKey)
}

// =========================================================================
// 7.4 异常容错与冲正补偿 Failure Recovery (全额回滚与指数退避重试)
// =========================================================================

// triggerRecovery 触发冲正补偿：先进行即时冲正，若失败入队后台异步退避重试
func triggerRecovery(ctx http_context.IHttpContext, cache resources.ICache, creditKey, balanceKey, preDeductKey string, deductType int, creditID string, estCost float64, userID string) {
	if cache == nil || estCost <= 0 {
		return
	}
	ok := executeRecovery(ctx.Context(), cache, creditKey, balanceKey, preDeductKey, deductType, creditID, estCost, userID)
	if !ok {
		log.Warnf("[dynamic-billing] instant recovery failed for user %s, enqueuing to recovery manager", userID)
		defaultRecoveryManager.Enqueue(&RecoveryTask{
			Cache:      cache,
			CreditKey:  creditKey,
			BalanceKey: balanceKey,
			LockKey:    preDeductKey,
			DeductType: deductType,
			CreditID:   creditID,
			EstCost:    estCost,
			UserID:     userID,
			MaxRetries: 5,
		})
	}
}

// executeRecovery 执行单次冲正原子 Lua：全额退还原预扣额度并删除预扣快照
func executeRecovery(ctx context.Context, cache resources.ICache, creditKey, balanceKey, preDeductKey string, deductType int, creditID string, estCost float64, userID string) bool {
	if cache == nil {
		return false
	}
	recoveryLua := `
		local creditKey = KEYS[1]
		local balanceKey = KEYS[2]
		local preKey = KEYS[3]
		local fallbackAmt = tonumber(ARGV[1])

		-- 读取并原子删除快照，防止重复冲正
		local snap = redis.call('hmget', preKey, 'deduct_type', 'credit_id', 'amount', 'credit_amount', 'balance_amount')
		redis.call('del', preKey)

		local dType = tonumber(snap[1] or ARGV[2])
		local cid = snap[2] or ARGV[3]
		local amt = tonumber(snap[3] or ARGV[1])
		local creditAmt = tonumber(snap[4] or "0")
		local balanceAmt = tonumber(snap[5] or "0")

		if dType == 3 then
			-- 跨账本组合预扣：分别原路退回代金券与现金
			if creditAmt > 0 and cid ~= "" then
				redis.call('hincrby', creditKey, cid, creditAmt)
			end
			if balanceAmt > 0 then
				redis.call('incrby', balanceKey, balanceAmt)
			end
			return 1
		elseif dType == 1 and cid ~= "" then
			local refund = creditAmt
			if refund <= 0 then refund = amt end
			if refund > 0 then
				redis.call('hincrby', creditKey, cid, refund)
			end
			return 1
		else
			local refund = balanceAmt
			if refund <= 0 then refund = amt end
			if refund > 0 then
				redis.call('incrby', balanceKey, refund)
			end
			return 1
		end
	`
	amt := int64(math.Round(estCost * 1000000))
	res := cache.Run(ctx, recoveryLua, []string{creditKey, balanceKey, preDeductKey}, amt, deductType, creditID)
	if _, err := res.Result(); err != nil {
		log.Errorf("[dynamic-billing] recovery redis error for user %s, key=%s: %v", userID, preDeductKey, err)
		return false
	}
	log.Infof("[dynamic-billing] successfully recovered pre-deduct %f for user %s, key=%s", estCost, userID, preDeductKey)
	return true
}
