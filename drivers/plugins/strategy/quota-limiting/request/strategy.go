package request

import (
	"errors"
	"fmt"
	quota_limiting "github.com/eolinker/apinto/drivers/plugins/strategy/quota-limiting"
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
	isProviderTenant := ctx.GetLabel("tenant") == ctx.GetLabel("provider_tenant")
	// 3. 依次对配额策略进行计数累加与超限判断
	for _, tss := range tenantStrategies {
		for _, st := range tss.Strategies() {
			if st.TargetType() == "channel" && isProviderTenant {
				continue
			}
			threshold := st.Threshold()
			if threshold <= 0 {
				continue
			}
			
			key, ttl := quota_limiting.BuildQuotaKeyAndTTL(ctx, s.key, st, now, "request")
			
			// 进行原子递增 1
			val, err := cache.IncrBy(ctx.Context(), key, 1, ttl).Result()
			if err != nil {
				log.Errorf("quota limiting incr error: %v, key: %s", err, key)
				continue
			}
			
			executedKeys = append(executedKeys, key)
			executedTTLs = append(executedTTLs, ttl)
			
			if val > int64(threshold) {
				// 超出配额阈值，回滚前面所有步骤已增加的计数
				for i, k := range executedKeys {
					cache.DecrBy(ctx.Context(), k, 1, executedTTLs[i])
				}
				
				// 进行 HTTP 拦截处理
				if httpContext, httpErr := http_service.Assert(ctx); httpErr == nil {
					// 将限制的信息写在header
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Id", st.Name())
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Type", "request")
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Period", st.Period().String())
					ctx.Response().SetHeader("X-Quota-Limiting-Strategy-Threshold", fmt.Sprintf("%f", st.Threshold()))
					ctx.WithValue("is_block", true)
					ctx.SetLabel("handler", "quota-limiting-request")
					if st.Response() != nil {
						st.Response().Response(ctx)
						return ErrQuotaExceeded
					}
					resp := httpContext.Response()
					resp.SetStatus(429, "429")
					resp.SetHeader("Content-Type", "application/json; charset=utf-8")
					resp.SetBody([]byte(`{"code":429,"message":"The throttling strategy has been triggered. Please check the throttling quota list in the call statistics of the console."}`))
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
