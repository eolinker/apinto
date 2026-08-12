package request

import (
	"errors"
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
	ErrQuotaExceeded = errors.New("refuse by request quota limiting strategy")
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
	// 1. 获取针对当前 Context 匹配的请求配额策略列表
	tenantStrategies, has := quota_limiting_strategy.GetRequestStrategies(ctx)
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
	executedKeys := make([]string, 0, len(tenantStrategies))
	executedTTLs := make([]time.Duration, 0, len(tenantStrategies))
	
	// 3. 依次对配额策略进行计数累加与超限判断
	for _, tss := range tenantStrategies {
		for _, st := range tss.Strategies() {
			threshold := st.Threshold()
			if threshold <= 0 {
				continue
			}
			
			key, ttl := s.buildQuotaKeyAndTTL(ctx, st, now)
			
			// 进行原子递增 1
			val, err := cache.IncrBy(ctx.Context(), key, 1, ttl).Result()
			if err != nil {
				log.Errorf("quota limiting incr error: %v, key: %s", err, key)
				continue
			}
			
			executedKeys = append(executedKeys, key)
			executedTTLs = append(executedTTLs, ttl)
			
			if val > threshold {
				// 超出配额阈值，回滚前面所有步骤已增加的计数
				for i, k := range executedKeys {
					cache.DecrBy(ctx.Context(), k, 1, executedTTLs[i])
				}
				
				// 进行 HTTP 拦截处理
				if httpContext, httpErr := http_service.Assert(ctx); httpErr == nil {
					ctx.WithValue("is_block", true)
					ctx.SetLabel("handler", "quota-limiting-request")
					if st.Response() != nil {
						st.Response().Response(ctx)
						return ErrQuotaExceeded
					}
					resp := httpContext.Response()
					resp.SetStatus(429, "429")
					resp.SetHeader("Content-Type", "application/json; charset=utf-8")
					resp.SetBody([]byte(`{"code":429,"message":"Request quota limit exceeded"}`))
				}
				return ErrQuotaExceeded
			}
		}
		
	}
	
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
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
	return nil
}

func (s *Strategy) Stop() error {
	return nil
}

func (s *Strategy) CheckSkill(skill string) bool {
	return eoscContext.FilterSkillName == skill
}
