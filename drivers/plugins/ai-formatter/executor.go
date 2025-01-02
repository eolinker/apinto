package ai_formatter

import (
	"errors"
	"fmt"
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
	provider string
	keyPool  convert.IKeyPool
	balances scope_manager.IProxyOutput[convert.IKeyPool]
	model    string
	modelCfg string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) doBalance(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain) error {
	if e.balances == nil {
		e.balances = scope_manager.Get[convert.IKeyPool]("ai_keys")
	}

	list := e.balances.List()
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
			ctx.SetProxy(originProxy)
			err := converter.RequestConvert(ctx, extender)
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
	return nil
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	if e.keyPool == nil {
		err := e.reset()
		if err != nil {
			return err
		}
	}
	cloneProxy := ctx.ProxyClone()
	convert.SetAIProvider(ctx, e.provider)
	convert.SetAIModel(ctx, e.model)
	selector := e.keyPool.Selector()
	var extender map[string]interface{}
	var err error
	for {
		resource, has := selector.Next()
		if !has {
			err = fmt.Errorf("no key resource for provider %s", e.provider)
			ctx.Response().SetStatus(402, "Payment Required")
			ctx.Response().SetBody([]byte(err.Error()))
			return errors.New("no key resource")
		}
		if extender == nil {
			fn, has := resource.ConverterDriver().GetModel(e.model)
			if !has {
				err := fmt.Errorf("model %s not found", e.model)
				ctx.Response().SetBody([]byte(err.Error()))
				ctx.Response().SetStatus(403, "Forbidden")
				return err
			}
			extender, err = fn(e.modelCfg)
			if err != nil {
				ctx.Response().SetBody([]byte(err.Error()))
				ctx.Response().SetStatus(403, "Forbidden")
				return err
			}
		}

		converter, has := resource.ConverterDriver().GetConverter(e.model)
		if !has {
			return errors.New("invalid model")
		}

		err = converter.RequestConvert(ctx, extender)
		if err != nil {
			continue
		}
		if next != nil {
			err = next.DoChain(ctx)
			if err != nil {
				// 转发失败，马上返回，一般和状态码相关
				if ctx.Response().StatusCode() == 504 {
					e.keyPool.Down()
					break
				}
				return err
			}
		}
		err = converter.ResponseConvert(ctx)
		if err != nil {
			return err
		}
		switch convert.GetAIStatus(ctx) {
		case convert.StatusInvalidRequest, convert.StatusNormal:
			return nil
		case convert.StatusInvalid, convert.StatusExpired, convert.StatusQuotaExhausted:
		case convert.StatusExceeded:
			resource.Down()
		}
	}
	return e.doBalance(ctx, cloneProxy, next)
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
	if workerResources == nil {
		return errors.New("workers not init")

	}
	v, has := workerResources.Get(e.provider)
	if !has {
		return errors.New("provider not found")
	}
	keyPool, ok := v.(convert.IKeyPool)
	if !ok {
		return errors.New("provider not implement IConverterDriver")
	}
	e.keyPool = keyPool

	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
