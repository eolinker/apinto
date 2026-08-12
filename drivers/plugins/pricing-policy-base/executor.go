package pricing_policy_base

import (
	"errors"
	"github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eocontext.IFilter = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	cacheKey   string
	defaultKey context_label.IKeyGenerator
	cal        price_calcular.ICalculator
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return errors.New("invalid config type")
	}
	return e.reset(cfg)
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) Destroy() {
	if e.cacheKey != "" {
		price_calcular.DelCalculator(e.cacheKey)
	}
	e.cal = nil
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	api := ctx.GetLabel("api")
	if api != "" {
		if e.cal != nil {
			price_calcular.SetICalculator(ctx, e.cal)
		} else {
			if e.defaultKey != nil {
				key := e.defaultKey.Key(ctx)
				cal, has := price_calcular.GetCalculator(key)
				if has {
					price_calcular.SetICalculator(ctx, cal)
				}
			}
		}
		
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (e *executor) reset(conf *Config) error {
	if conf.DefaultCalculatorKey != "" {
		e.defaultKey = context_label.NewKeyGenerator(conf.DefaultCalculatorKey)
	}
	if conf.Currency != "" && conf.ContextVariables != nil && conf.AdvancedRules != nil {
		cal, err := price_calcular.NewCalculator(e.Id(), conf.Currency, conf.ContextVariables, conf.AdvancedRules)
		if err != nil {
			return err
		}
		e.cal = cal
		if conf.CacheKey != "" {
			price_calcular.SetCalculator(conf.CacheKey, cal)
		}
	}
	e.cacheKey = conf.CacheKey
	
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
