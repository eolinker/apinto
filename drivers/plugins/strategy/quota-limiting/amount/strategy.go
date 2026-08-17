package amount

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	quota_limiting "github.com/eolinker/apinto/drivers/plugins/strategy/quota-limiting"
	price_calcular "github.com/eolinker/apinto/price-calcular"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var (
	ErrQuotaExceeded = errors.New("refuse by amount quota limiting strategy")
)

const amountUnit = 1000000 // 1 元 = 1000000 微单位

type Strategy struct {
	drivers.WorkerBase
	key                  context_label.IKeyGenerator
	redisID              string
	priceKey             context_label.IKeyGenerator
	versionKey           context_label.IKeyGenerator
	bindResourceGroupKey context_label.IKeyGenerator
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	return http_service.DoHttpFilter(s, ctx, next)
}

func (s *Strategy) DoHttpFilter(ctx http_service.IHttpContext, next eoscContext.IChain) error {
	start := time.Now()
	defer func() {
		log.Info("quota-limiting spend time: ", time.Now().Sub(start))
	}()
	api := ctx.GetLabel("api")
	if api == "" {
		return nil
	}

	calcId := fmt.Sprintf("%s:%s", ctx.GetLabel("resource_type"), ctx.GetLabel("resource"))
	if calcId == "" {
		ctx.Response().SetStatus(500, "Internal Server Error")
		ctx.Response().SetBody([]byte("calculator id not found in context_label"))
		return fmt.Errorf("calculator id not found in context_label, api: %s", api)
	}
	calc, ok := price_calcular.GetCalculator(calcId)
	if !ok {
		ctx.Response().SetStatus(500, "Internal Server Error")
		ctx.Response().SetBody([]byte("calculator not found"))
		return fmt.Errorf("calculator not found, api: %s, calculator id: %s", api, calcId)
	}

	// 2. 获取 Cache 存储引擎（Redis 或 本地 Cache）
	var cache resources.ICache
	if s.redisID != "" {
		caches := scope_manager.Auto[resources.ICache](s.redisID, "redis").List()
		if len(caches) > 0 {
			cache = caches[0]
		}
	}
	if cache == nil {
		// 价钱相关的必须要配置redis，不考虑本地存储
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	getPricingDataStart := time.Now()
	pricingData, err := s.getPricingData(ctx, cache)
	log.Info("get pricing data time: ", time.Since(getPricingDataStart))
	if err != nil || pricingData == nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	context_label.SetPriceVersion(ctx, pricingData.BasicInfo.Version)
	quota_limiting_strategy.SetUserOfResourceGroup(ctx, pricingData.BasicInfo.ResourceGroupID)
	getStrategiesStart := time.Now()
	// 1. 获取针对当前 Context 匹配的金额配额策略列表
	tenantStrategies, has := quota_limiting_strategy.GetAmountStrategies(ctx)
	log.Info("get strategies time: ", time.Since(getStrategiesStart))
	if !has || len(tenantStrategies) == 0 {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	getPricingDataMapStart := time.Now()
	pm, err := getPricingDataMap(ctx, cache, s.versionKey)
	log.Info("get pricing data map time: ", time.Since(getPricingDataMapStart))
	if err != nil {
		log.Errorf("get pricing data map error: %v", err)
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	pm[ctx.GetLabel("tenant")] = pricingData
	now := time.Now()

	executedItems := make([]*executedItem, 0, len(tenantStrategies))
	isProviderTenant := ctx.GetLabel("tenant") == ctx.GetLabel("provider_tenant")
	// 3. 【预扣阶段 Pre-deduct】：按策略预先增加金额计数并检查是否超出阈值
	for _, tss := range tenantStrategies {
		p, ok := pm[tss.Tenant()]
		if !ok {
			continue
		}
		preDeductStart := time.Now()
		amount, _, err := calc.PreDeduct(ctx, p)
		log.Info("pre deduct calculation time: ", time.Since(preDeductStart))
		if err != nil {
			log.Errorf("amount quota limiting pre-deduct error: %v, tenant: %s", err, tss.Tenant())
			continue
		}

		traversalStrategyStart := time.Now()
		amountInt := int64(math.Round(amount * amountUnit))
		for _, st := range tss.Strategies() {
			if st.TargetType() == "channel" && isProviderTenant {
				continue
			}
			threshold := st.Threshold()
			if threshold <= 0 {
				continue
			}
			thresholdInt := int64(math.Round(threshold * amountUnit))

			key, ttl := quota_limiting.BuildQuotaKeyAndTTL(ctx, s.key, st, now, "amount")

			// 原子执行预扣 (IncrBy estimateAmountInt)
			incrStart := time.Now()
			val, incrErr := cache.IncrBy(ctx.Context(), key, amountInt, ttl).Result()
			log.Info("redis pre-deduct incr time: ", time.Since(incrStart))
			if incrErr != nil {
				log.Errorf("amount quota limiting pre-deduct incr error: %v, key: %s", incrErr, key)
				continue
			}
			executedItems = append(executedItems, &executedItem{
				key:           key,
				pricing:       p,
				preDecrAmount: amountInt,
				ttl:           ttl,
			})

			// 检查超限：如果预扣后的累计金额超过配额阈值 thresholdInt
			if val > thresholdInt {
				// 【回滚阶段 Rollback】：退还前面已对其他策略预扣的金额数量
				rollbackStart := time.Now()
				for _, k := range executedItems {
					_, decrErr := cache.DecrBy(ctx.Context(), k.key, k.preDecrAmount, k.ttl).Result()
					if decrErr != nil {
						log.Errorf("amount quota limiting rollback error: %v, key: %s", decrErr, k.key)
					}
				}
				log.Info("rollback time: ", time.Since(rollbackStart))

				// 进行 HTTP 拦截处理并返回 429
				if httpContext, httpErr := http_service.Assert(ctx); httpErr == nil {
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Id", st.Name())
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Type", "amount")
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Period", st.Period().String())
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Threshold", fmt.Sprintf("%f", st.Threshold()))
					ctx.WithValue("is_block", true)
					ctx.SetLabel("handler", "quota-limiting-amount")
					if st.Response() != nil {
						st.Response().Response(ctx)
						return ErrQuotaExceeded
					}
					resp := httpContext.Response()
					resp.SetStatus(429, "429")
					resp.SetHeader("Content-Type", "application/json; charset=utf-8")
					resp.SetBody([]byte(`{"code":429,"The 'amount' quota strategy has been triggered. Please check the 'amount' quota list in the call statistics of the console."}`))
				}
				return ErrQuotaExceeded
			}
		}
		log.Infof("finish traversal strategy: %s, time: %v", tss.Tenant(), time.Since(traversalStrategyStart))
	}
	ctx.Proxy().AppendBodyFinish(func(ctx http_service.IHttpContext) {
		settleStart := time.Now()
		settle(ctx, cache, calc, executedItems)
		log.Info("settle body finish time: ", time.Since(settleStart))
	})

	// 4. 执行后续处理链
	if next != nil {
		doChainStart := time.Now()
		err = next.DoChain(ctx)
		log.Info("do chain (next filters and proxy) time: ", time.Since(doChainStart))
		if err != nil {
			settle(ctx, cache, calc, executedItems)
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		// 如果是流式，则直接返回
		return nil
	}
	settleStart := time.Now()
	settle(ctx, cache, calc, executedItems)
	log.Info("settle time: ", time.Since(settleStart))
	return nil
}

type executedItem struct {
	key           string
	pricing       *price_calcular.PricingData
	preDecrAmount int64
	ttl           time.Duration
}

func settle(ctx http_service.IHttpContext, cache resources.ICache, calc price_calcular.ICalculator, settleAmounts []*executedItem) {
	for _, s := range settleAmounts {
		// 5.1 【补扣与结算阶段 Post-settle / Rollback】
		result, err := calc.CalculateByRule(ctx, s.pricing)
		if err != nil {
			log.Errorf("amount quota limiting post-settle calculate error: %v, key: %s", err, s.key)
			continue
		}

		actualAmountInt := int64(math.Round(result.Sale * amountUnit))

		// 请求正常成功且计算出实际金额消费数，进行补扣/多退少补结算 (diff = actualAmountInt - preDecrAmount[i])
		diff := actualAmountInt - s.preDecrAmount
		if diff != 0 {
			_, decrErr := cache.DecrBy(ctx.Context(), s.key, -diff, s.ttl).Result()
			if decrErr != nil {
				log.Errorf("amount quota limiting post-settle error: %v, key: %s", decrErr, s.key)
			}
		}
	}
}

func (s *Strategy) Destroy() {
}

func (s *Strategy) Start() error {
	return nil
}

func (s *Strategy) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	s.redisID = string(cfg.Cache)
	s.key = context_label.NewKeyGenerator(cfg.Key)
	s.versionKey = context_label.NewKeyGenerator(cfg.PriceKey)
	return nil
}

func (s *Strategy) Stop() error {
	return nil
}

func (s *Strategy) CheckSkill(skill string) bool {
	return eoscContext.FilterSkillName == skill
}

func getPricingDataMap(ctx eoscContext.EoContext, cache resources.ICache, priceKey context_label.IKeyGenerator) (map[string]*price_calcular.PricingData, error) {
	result := make(map[string]*price_calcular.PricingData)
	var fn func(ver string) error
	fn = func(ver string) error {
		if ver == "" {
			return nil
		}
		key := priceKey.Key(ctx, func(ctx eoscContext.EoContext, label string) string {
			if label == "version" {
				return ver
			}
			return ctx.GetLabel(label)
		})
		data, err := cache.Get(ctx.Context(), key).Result()
		if err != nil {
			return fmt.Errorf("get pricing data from cache error: %v, key: %s", err, key)
		}
		var p price_calcular.PricingData
		err = json.Unmarshal([]byte(data), &p)
		if err != nil {
			return fmt.Errorf("unmarshal pricing data error: %v, key: %s", err, key)
		}
		result[p.BasicInfo.TenantID] = &p
		return fn(p.BasicInfo.Rely)
	}
	err := fn(context_label.GetPriceVersion(ctx))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Strategy) getPricingData(ctx http_service.IHttpContext, cache resources.ICache) (*price_calcular.PricingData, error) {
	var priceData *price_calcular.PricingData
	isUser := context_label.IsUserConsumer(ctx)
	if !isUser {
		priceKey := s.priceKey.Key(ctx)

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
		key := s.bindResourceGroupKey.Key(ctx)
		rgs, ok := customerVar.GetAll(key)
		if !ok {
			return nil, fmt.Errorf("{\"error\":\"no resource group found\"}")
		}
		pds := make([]*price_calcular.PricingData, 0, len(rgs))
		for key := range rgs {
			priceKey := s.priceKey.Key(ctx, func(ctx eoscContext.EoContext, label string) string {
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
	if priceData != nil {
		if priceData.BasicInfo != nil {
			context_label.SetPriceVersion(ctx, priceData.BasicInfo.Version)
			context_label.SetPriceRelyVersion(ctx, priceData.BasicInfo.Rely)
		}
	} else {
		return nil, fmt.Errorf("{\"error\":\"pricing data not found\"}")
	}
	return priceData, nil
}
