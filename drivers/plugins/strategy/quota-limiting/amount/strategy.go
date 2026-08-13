package amount

import (
	"encoding/json"
	"errors"
	"fmt"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"math"
	"time"
	
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
	bindResourceGroupKey context_label.IKeyGenerator
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	return http_service.DoHttpFilter(s, ctx, next)
}

func (s *Strategy) DoHttpFilter(ctx http_service.IHttpContext, next eoscContext.IChain) error {
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
	// 1. 获取针对当前 Context 匹配的金额配额策略列表
	tenantStrategies, has := quota_limiting_strategy.GetAmountStrategies(ctx)
	if !has || len(tenantStrategies) == 0 {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	pricingData, err := s.getPricingData(ctx, cache)
	if err != nil || pricingData == nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	
	pm, err := getPricingDataMap(ctx, cache, s.priceKey)
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
	
	//estimateAmountInt := int64(math.Round(estimateAmount * amountUnit))
	
	// 3. 【预扣阶段 Pre-deduct】：按策略预先增加金额计数并检查是否超出阈值
	for _, tss := range tenantStrategies {
		p, ok := pm[tss.Tenant()]
		if !ok {
			continue
		}
		amount, _, err := calc.PreDeduct(ctx, p)
		if err != nil {
			log.Errorf("amount quota limiting pre-deduct error: %v, tenant: %s", err, tss.Tenant())
			continue
		}
		amountInt := int64(math.Round(amount * amountUnit))
		for _, st := range tss.Strategies() {
			threshold := st.Threshold()
			if threshold <= 0 {
				continue
			}
			thresholdInt := threshold * amountUnit
			
			key, ttl := s.buildQuotaKeyAndTTL(ctx, st, now)
			
			// 原子执行预扣 (IncrBy estimateAmountInt)
			val, incrErr := cache.IncrBy(ctx.Context(), key, amountInt, ttl).Result()
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
				for _, k := range executedItems {
					_, decrErr := cache.DecrBy(ctx.Context(), k.key, k.preDecrAmount, k.ttl).Result()
					if decrErr != nil {
						log.Errorf("amount quota limiting rollback error: %v, key: %s", decrErr, k.key)
					}
				}
				
				// 进行 HTTP 拦截处理并返回 429
				if httpContext, httpErr := http_service.Assert(ctx); httpErr == nil {
					ctx.WithValue("is_block", true)
					ctx.SetLabel("handler", "quota-limiting-amount")
					if st.Response() != nil {
						st.Response().Response(ctx)
						return ErrQuotaExceeded
					}
					resp := httpContext.Response()
					resp.SetStatus(429, "429")
					resp.SetHeader("Content-Type", "application/json; charset=utf-8")
					resp.SetBody([]byte(`{"code":429,"message":"Amount quota limit exceeded"}`))
				}
				return ErrQuotaExceeded
			}
		}
	}
	ctx.Proxy().AppendBodyFinish(func(ctx http_service.IHttpContext) {
		settle(ctx, cache, calc, executedItems)
	})
	
	// 4. 执行后续处理链
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		// 如果是流式，则直接返回
		return nil
	}
	settle(ctx, cache, calc, executedItems)
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

// buildQuotaKeyAndTTL 按照技术方案构造标准 Key 结构及 TTL
// 规范 Key 结构: {product}:quota-limiting:{策略uuid}:{调用方类型}:{调用方uuid}:{配额维度}:{时间}
func (s *Strategy) buildQuotaKeyAndTTL(ctx eoscContext.EoContext, st quota_limiting_strategy.IStrategy, now time.Time) (string, time.Duration) {
	var ttl time.Duration
	
	return s.key.Key(ctx, func(ctx eoscContext.EoContext, label string) string {
		switch label {
		case "target_type":
			return st.TargetType()
		case "strategy":
			return st.ID()
		case "period":
			return st.Period().String()
		case "time_format":
			var timeStr string
			switch st.Period() {
			case quota_limiting_strategy.PeriodSecond:
				timeStr = now.Format("20060102150405")
				ttl = 2 * time.Second
			case quota_limiting_strategy.PeriodMinute:
				timeStr = now.Format("200601021504")
				ttl = time.Duration(60-now.Second())*time.Second + 10*time.Second
			case quota_limiting_strategy.PeriodHour:
				timeStr = now.Format("2006010215")
				ttl = time.Duration(3600-now.Minute()*60-now.Second())*time.Second + 60*time.Second
			case quota_limiting_strategy.PeriodDay:
				timeStr = now.Format("20060102")
				ttl = time.Duration(86400-now.Hour()*3600-now.Minute()*60-now.Second())*time.Second + 300*time.Second
			case quota_limiting_strategy.PeriodMonth:
				timeStr = now.Format("200601")
				nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
				ttl = nextMonth.Sub(now) + 3600*time.Second
			case quota_limiting_strategy.PeriodTotal:
				timeStr = "total"
				ttl = -1
			default:
				timeStr = now.Format("20060102150405")
				ttl = 2 * time.Second
			}
			return timeStr
		}
		return ctx.GetLabel(Name)
	}), ttl
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
	s.priceKey = context_label.NewKeyGenerator(cfg.PriceKey)
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
