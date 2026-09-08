package failover_strategy

import (
	"time"

	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuatorSet = newtActuator()
}

type ActuatorSet interface {
	strategy.IStrategyHandler
	Set(id string, val *FailoverHandler)
	Del(id string)
	Extract(ctx http_service.IHttpContext) *FailoverHandler
}

type tActuator struct {
	extractor IExtractor
}

func newtActuator() *tActuator {
	return &tActuator{
		extractor: NewExtractor(),
	}
}

func (a *tActuator) Set(id string, val *FailoverHandler) {
	a.extractor.Set(id, val)
}

func (a *tActuator) Del(id string) {
	a.extractor.Del(id)
}

func (a *tActuator) Extract(ctx http_service.IHttpContext) *FailoverHandler {
	return a.extractor.Extract(ctx)
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}

	// 通过提取器获取命中策略（若请求 resource 在范围则命中，否则寻找 provider 匹配策略）
	matchedHandler := a.extractor.Extract(httpCtx)
	if matchedHandler == nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	// 1. 直接切换模式 (direct)
	if matchedHandler.TriggerType() == TriggerTypeDirect {
		matchedHandler.ApplyFailover(httpCtx, 0)
		ctx.WithValue("is_block", true)
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	// 2. 条件触发模式 (condition)
	cloneProxy := httpCtx.ProxyClone()
	start := time.Now()
	if next != nil {
		err = next.DoChain(ctx)
	}
	cost := time.Since(start)

	if matchedHandler.IsTriggerCondition(httpCtx, err, cost) {
		providers := matchedHandler.Providers()
		for i := range providers {
			matchedHandler.ApplyFailover(httpCtx, i)
			ctx.WithValue("is_block", true)
			if cloneProxy != nil {
				httpCtx.SetProxy(cloneProxy)
			}
			if next != nil {
				retryStart := time.Now()
				retryErr := next.DoChain(ctx)
				if !matchedHandler.IsTriggerCondition(httpCtx, retryErr, time.Since(retryStart)) {
					return nil
				}
				err = retryErr
			}
		}
	}

	return err
}

type handlerListSort []*FailoverHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {
	return hs[i].Priority() < hs[j].Priority()
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}

// Extract 从上下文提取匹配的灾备策略Handler（先匹配 resource，未命中再匹配 provider，最多一个）
func Extract(ctx http_service.IHttpContext) *FailoverHandler {
	return actuatorSet.Extract(ctx)
}

// GetStrategy 从上下文获取匹配的灾备策略Handler
func GetStrategy(ctx http_service.IHttpContext) (*FailoverHandler, bool) {
	h := actuatorSet.Extract(ctx)
	return h, h != nil
}
