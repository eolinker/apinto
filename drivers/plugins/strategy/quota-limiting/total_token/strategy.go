package total_token

import (
	"errors"
	"fmt"
	price_calcular "github.com/eolinker/apinto/price-calcular"
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
	ErrQuotaExceeded = errors.New("refuse by total token quota limiting strategy")
)

type Strategy struct {
	drivers.WorkerBase
	key     context_label.IKeyGenerator
	redisID string
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
	rules, err := calc.Rules(ctx)
	if err != nil {
		ctx.Response().SetStatus(500, "Internal Server Error")
		ctx.Response().SetBody([]byte("calculator rules error"))
		return fmt.Errorf("calculator rules error, api: %s, calculator id: %s, error: %v", api, calcId, err)
	}
	estimateToken := price_calcular.GetPreTotalToken(ctx, rules)
	// 优先从 context_label 中获取预扣的总 Token 数
	if estimateToken <= 0 {
		// 预扣Token值没设置，不考虑执行，快速返回
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	// 1. 获取针对当前 Context 匹配的总 Token 配额策略列表
	tenantStrategies, has := quota_limiting_strategy.GetTotalTokenStrategies(ctx)
	if !has || len(tenantStrategies) == 0 {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
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
		cache = resources.LocalCache()
	}
	
	now := time.Now()
	items := make([]*executedItem, 0, len(tenantStrategies))
	
	// 3. 【预扣阶段 Pre-deduct】：按策略预先增加 Token 计数并检查是否超出阈值
	for _, tss := range tenantStrategies {
		for _, st := range tss.Strategies() {
			threshold := st.Threshold()
			if threshold <= 0 {
				continue
			}
			
			key, ttl := s.buildQuotaKeyAndTTL(ctx, st, now)
			
			// 原子执行预扣 (IncrBy estimateToken)
			val, incrErr := cache.IncrBy(ctx.Context(), key, estimateToken, ttl).Result()
			if incrErr != nil {
				log.Errorf("total token quota limiting pre-deduct incr error: %v, key: %s", incrErr, key)
				continue
			}
			items = append(items, &executedItem{
				key: key,
				ttl: ttl,
			})
			
			// 检查超限：如果预扣后的累计 Token 超过配额阈值 threshold
			if val > threshold {
				// 【回滚阶段 Rollback】：退还前面已对其他策略预扣的 Token 数量
				for _, k := range items {
					_, decrErr := cache.DecrBy(ctx.Context(), k.key, estimateToken, k.ttl).Result()
					if decrErr != nil {
						log.Errorf("total token quota limiting rollback error: %v, key: %s", decrErr, k)
					}
				}
				
				// 进行 HTTP 拦截处理并返回 429
				if httpContext, httpErr := http_service.Assert(ctx); httpErr == nil {
					ctx.WithValue("is_block", true)
					ctx.SetLabel("handler", "quota-limiting-total-token")
					if st.Response() != nil {
						st.Response().Response(ctx)
						return ErrQuotaExceeded
					}
					resp := httpContext.Response()
					resp.SetStatus(429, "429")
					resp.SetHeader("Content-Type", "application/json; charset=utf-8")
					resp.SetBody([]byte(`{"code":429,"message":"Total token quota limit exceeded"}`))
				}
				return ErrQuotaExceeded
			}
		}
	}
	
	// 4. 执行后续处理链
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	settle(ctx, cache, items, estimateToken)
	return nil
}

type executedItem struct {
	key string
	ttl time.Duration
}

func settle(ctx http_service.IHttpContext, cache resources.ICache, items []*executedItem, estimateToken int64) {
	// 5. 【补扣与结算阶段 Post-settle / Rollback】
	// 通过 context_label 统一接口获取请求完成后实际消耗的总 Token 数量
	actualToken := context_label.GetActualTotalToken(ctx)
	
	// 如果请求执行发生异常/报错，或者未产生有效 Token，则退还（回滚）预扣额度
	if actualToken <= 0 {
		for _, k := range items {
			_, decrErr := cache.DecrBy(ctx.Context(), k.key, estimateToken, k.ttl).Result()
			if decrErr != nil {
				log.Errorf("total token quota limiting refund pre-deduct error: %v, key: %s", decrErr, k)
			}
		}
	}
	
	// 请求正常成功且计算出实际 Token 消费数，进行补扣/多退少补结算 (diff = actualToken - estimateToken)
	diff := int64(actualToken) - estimateToken
	if diff != 0 {
		for _, k := range items {
			if diff > 0 {
				// 实际消费多于预扣：补扣多出的差额
				_, settleErr := cache.IncrBy(ctx.Context(), k.key, diff, k.ttl).Result()
				if settleErr != nil {
					log.Errorf("total token quota limiting settle diff incr error: %v, key: %s", settleErr, k)
				}
			} else {
				// 实际消费少于预扣：退还多预扣的差额
				_, settleErr := cache.DecrBy(ctx.Context(), k.key, -diff, k.ttl).Result()
				if settleErr != nil {
					log.Errorf("total token quota limiting settle diff decr error: %v, key: %s", settleErr, k)
				}
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
	return nil
}

func (s *Strategy) Stop() error {
	return nil
}

func (s *Strategy) CheckSkill(skill string) bool {
	return eoscContext.FilterSkillName == skill
}
