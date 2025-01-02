package ai_balance

import (
	"errors"
	"sort"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type executor struct {
	drivers.WorkerBase
	keyPools scope_manager.IProxyOutput[convert.IKeyPool]
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	if e.keyPools == nil {
		e.keyPools = scope_manager.Get[convert.IKeyPool]("ai_keys")
	}
	if next != nil {
		err := next.DoChain(ctx)
		status := convert.GetAIStatus(ctx)
		if status == "" || status == convert.StatusNormal || status == convert.StatusInvalidRequest {
			// 不需要进行负载的场景
			return err
		}
		list := e.keyPools.List()
		sort.Slice(list, func(i, j int) bool {
			return list[i].Priority() < list[j].Priority()
		})
		for _, l := range list {
			if !l.Health() {
				continue
			}
			convert.SetAIProvider(ctx, l.Provider())
			convert.SetAIModel(ctx, l.Model())
			extender := l.ModelConfig()
			selector := l.Selector()
			for {
				resource, has := selector.Next()
				if !has {
					l.Down()
					// 轮询完当前供应商
					break
				}
				converter, has := resource.ConverterDriver().GetConverter(l.Model())
				if !has {
					l.Down()
					return errors.New("invalid model")
				}

				err = converter.RequestConvert(ctx, extender)
				if err != nil {
					continue
				}
				if next != nil {
					err = next.DoChain(ctx)
					if ctx.Response().StatusCode() == 504 {
						// 供应商响应超时，跳过当前供应商
						convert.SetAIStatusTimeout(ctx)
						l.Down()
						break
					}
					if err != nil {
						return err
					}
				}
				err = converter.ResponseConvert(ctx)
				if err != nil {
					return err
				}
				switch convert.GetAIStatus(ctx) {
				case convert.StatusInvalid, convert.StatusExpired, convert.StatusQuotaExhausted:
					resource.Down()
				case convert.StatusExceeded:
					// 熔断
					resource.Breaker()
				}
			}
		}
	}
	return nil
}

func (e *executor) Destroy() {
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset() error {
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
