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
	redisID                       string
	defaultConcurrencyLimit       int
	enableBalance                 bool
	taskKeyGenerator              context_label.IKeyGenerator
	concurrencyKeyGenerator       context_label.IKeyGenerator
	balanceKeyGenerator           context_label.IKeyGenerator
	priceKeyGenerator             context_label.IKeyGenerator
	bindResourceGroupKeyGenerator context_label.IKeyGenerator
}

// TaskInfo 描述了异步任务生成的元数据，用于二次状态查询时反查账户和资源组
type TaskInfo struct {
	ResourceID   string                      `json:"resource"`
	App          string                      `json:"app"`
	IsCharged    bool                        `json:"is_charged"`
	PreDeductKey string                      `json:"pre_deduct_key,omitempty"`
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
	e.enableBalance = cfg.EnableBalance
	e.balanceKeyGenerator = context_label.NewKeyGenerator(cfg.BalanceKey)
	e.priceKeyGenerator = context_label.NewKeyGenerator(cfg.PriceKey)
	e.taskKeyGenerator = context_label.NewKeyGenerator(cfg.TaskKey)
	e.concurrencyKeyGenerator = context_label.NewKeyGenerator(cfg.ConcurrencyKey)
	e.bindResourceGroupKeyGenerator = context_label.NewKeyGenerator(cfg.BindResourceGroupKey)
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

				// 还原预扣快照 key，供本次 query 阶段结算/回滚使用
				if info.PreDeductKey != "" {
					context_label.SetPreDeductKey(ctx, info.PreDeductKey)
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
	// ==========================================
	// 10. 组装资源定价价格 Key (优先使用固化的 PricingData，若无则从 Redis 动态读取)
	// ==========================================
	priceData, err := e.getPricingData(ctx, cache, snapshottedPricingData)
	if err != nil {
		ctx.Response().SetBody([]byte(err.Error()))
		ctx.Response().SetStatus(400, "Bad Request")
		return err
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
	if isPreCheckRequired && e.enableBalance && cache != nil && !hasFreePricePlan(priceData) {
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
	// 7.5 预扣逻辑（仅同步/Immediate 阶段）
	//
	// 高并发要点：
	//   a) 预扣时使用一条 Lua 脚本原子完成 "余额 DECRBY N" + "SETEX preKey N ttl"，
	//      保证余额扣了必然有预扣快照，反之亦然。
	//   b) 每次请求生成独立的 preKey（{prefix}:{app}:{request_id}），
	//      因此并发请求互不干扰。
	//   c) 结算/回滚也用一条 Lua 脚本：内部 GET preKey→pre; DEL preKey;
	//      DECRBY/INCRBY 余额相应差额。DEL 保证脚本天然幂等——
	//      同一请求多次触发结算/回滚，账面变化只会发生一次。
	// ==========================================
	var preDeductKey string
	if e.enableBalance && cache != nil && isPreCheckRequired && !hasFreePricePlan(priceData) {
		amt, rid, pdErr := calc.PreDeduct(ctx, priceData)
		if pdErr != nil {
			log.Errorf("[dynamic-billing] pre-deduct calculate error for resource %s: %v", resourceID, pdErr)
		} else if amt > 0 {
			// preDeductKey 模板中的 {request_id} 由 KeyGenerator 通过下方 GetLabelFunc 解析成
			// ctx.RequestId()，保证每次请求（含并发异步任务）都有独立的预扣快照 key。
			preDeductKey = preDeductKeyGenerator.Key(ctx, func(c eocontext.EoContext, label string) string {
				if label == "request_id" {
					return c.RequestId()
				}
				return ""
			})
			// 异步 TaskCreate 阶段：预扣快照需与 TaskInfo 缓存同寿（24h），
			// 避免 query 阶段到来时预扣快照已过期导致无法结算/回滚。
			ttl := preDeductTTL
			if billingMode == context_label.BillingModeTaskCreate {
				ttl = 24 * time.Hour
			}
			if !executePreDeduct(ctx.Context(), cache, balanceKey, preDeductKey, amt, ttl, app) {
				log.Errorf("[dynamic-billing] pre-deduct failed for user %s, amount=%f", app, amt)
				ctx.Response().SetStatus(http.StatusPaymentRequired, "402")
				ctx.Response().SetBody([]byte(`{"error":"insufficient balance"}`))
				return errors.New(`{"error":"insufficient balance"}`)
			}
			context_label.SetAmountPreDeduct(ctx, fmt.Sprintf("%f", amt))
			context_label.SetPreDeductKey(ctx, preDeductKey)
			log.DebugF("[dynamic-billing] pre-deducted %f for user %s on resource %s (rule=%s, key=%s)",
				amt, app, resourceID, rid, preDeductKey)
		}
	}

	//var res *pricing_policy.CalculateResult
	// 结算回调仅对 Immediate（同步）模式注册：
	//   - TaskCreate 阶段的响应体是任务创建结果，此时用它计算实际用量是错误的；
	//     真正的用量在 TaskQuery 阶段通过 settleBilling 结算。
	//   - TaskQuery 阶段的结算已经统一走 settleBilling → settlePreDeduct/refundPreDeduct，
	//     若此处再注册 AppendBodyFinish 会造成双重扣款。
	if billingMode == context_label.BillingModeImmediate && ctx.GetLabel("resource_type") == "api" {
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			res, err := calc.Calculate(ctx, e.enableBalance, priceData)
			if err != nil {
				log.Errorf("[dynamic-billing] calculate error: %v", err)
				// 计算失败：整额退还预扣
				refundPreDeduct(ctx, cache, balanceKey, preDeductKey, app)
				return
			}
			context_label.SetAmountCost(ctx, fmt.Sprintf("%f", res.Cost))
			context_label.SetAmountSale(ctx, fmt.Sprintf("%f", res.Sale))
			context_label.SetAmountOfficial(ctx, fmt.Sprintf("%f", res.Official))
			if e.enableBalance && cache != nil {
				settlePreDeduct(ctx, cache, balanceKey, preDeductKey, res.Sale, app)
			} else {
				log.DebugF("[dynamic-billing] balance settlement skipped for user %s on resource %s (rule=%s, key=%s)",
					app, resourceID, preDeductKey)
			}
		})
	} else if billingMode == context_label.BillingModeImmediate && ctx.GetLabel("resource_type") == "ai" {
		var res *price_calcular.CalculateResult
		var streamBuffer []byte
		ctx.Proxy().AppendBodyFinish(func(ctx http_context.IHttpContext) {
			if len(streamBuffer) > 0 {
				line := strings.TrimSpace(string(streamBuffer))
				if line != "" {
					tmp, err := calc.CalculateFromChunk(ctx, e.enableBalance, []byte(line), priceData)
					if err == nil && tmp != nil {
						res = tmp
					}
				}
				streamBuffer = nil
			}
			if res != nil && res.Sale > 0 {
				if e.enableBalance && cache != nil {
					settlePreDeduct(ctx, cache, balanceKey, preDeductKey, res.Sale, app)
				}
			} else {
				// 未完成实际结算：整额退还预扣
				refundPreDeduct(ctx, cache, balanceKey, preDeductKey, app)
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
			// 拼接到请求级流式缓冲区
			streamBuffer = append(streamBuffer, body...)

			// 查找最后一个换行符
			lastNL := bytes.LastIndexByte(streamBuffer, '\n')
			if lastNL == -1 {
				// 尚未接收到完整的行，等待下一个 Chunk
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
				tmp, err := calc.CalculateFromChunk(ctx, e.enableBalance, []byte(line), priceData)
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
	_ = preDeductKey // 预留：可用于埋点/日志时区分命中的规则

	// ==========================================
	// 8. 转发下游链路 (Do Chain Forwarding)
	// ==========================================
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			log.Errorf("[dynamic-billing] downstream chain forwarding error: %v", err)
			// 请求链路失败：退回已预扣的金额，避免用户在无实际服务下被扣款
			refundPreDeduct(ctx, cache, balanceKey, preDeductKey, app)
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

	// 同步结算非流式响应
	return e.settleBilling(ctx, calc, priceData, balanceKey, resourceID, app, cache, billingMode, isCharged)
}

func (e *executor) getPricingData(ctx http_context.IHttpContext, cache resources.ICache, snapPricingData *price_calcular.PricingData) (*price_calcular.PricingData, error) {
	var priceData *price_calcular.PricingData
	if e.enableBalance {
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
	} else if e.enableBalance && priceData == nil {
		return nil, fmt.Errorf("{\"error\":\"pricing data not found\"}")
	}
	return priceData, nil
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

	if e.enableBalance && cache != nil {
		// 若存在预扣快照，则以 Lua 幂等结算做多退少补；否则按 sale 直接扣款
		preKey := context_label.GetPreDeductKey(ctx)
		settlePreDeduct(ctx, cache, balanceKey, preKey, res.Sale, app)
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
		// TaskQuery 阶段结算/回滚 都需要 TaskCreate 阶段落下的预扣快照 key。
		// 它已在 DoHttpFilter 反查 TaskInfo 时被写回上下文，这里直接读取即可。
		preDeductKey := context_label.GetPreDeductKey(ctx)
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
			if e.enableBalance && cache != nil {
				// 任务执行成功：以预扣快照做多退少补结算；
				// 若无预扣快照（例如免费策略或 TaskCreate 阶段未预扣），
				// settlePreDeduct 会退化为按 res.Sale 直接扣款。
				settlePreDeduct(ctx, cache, balanceKey, preDeductKey, res.Sale, app)
			}
		case context_label.TaskStatusFailed:
			isCharged = true
			// 任务执行失败：整额退回 TaskCreate 阶段的预扣金额。
			if e.enableBalance && cache != nil {
				refundPreDeduct(ctx, cache, balanceKey, preDeductKey, app)
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
