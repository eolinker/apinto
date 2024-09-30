package ai_formatter

import (
	"errors"

	"github.com/eolinker/apinto/convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type executor struct {
	drivers.WorkerBase
	model     string
	extender  map[string]interface{}
	converter convert.IConverter
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	err := e.converter.RequestConvert(ctx, e.extender)
	if err != nil {
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	return e.converter.ResponseConvert(ctx)
}

func (e *executor) Destroy() {
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	w, ok := workers[cfg.Provider]
	if !ok {
		return errors.New("invalid provider")
	}
	if v, ok := w.(convert.IConverterDriver); ok {
		converter, has := v.GetConverter(cfg.Model)
		if !has {
			return errors.New("invalid model")
		}
		f, has := v.GetModel(cfg.Model)
		if !has {
			return errors.New("invalid model")
		}

		extender, err := f(cfg.Config)
		if err != nil {
			return err
		}
		e.converter = converter
		e.model = cfg.Model
		e.extender = extender
		return nil
	}
	return errors.New("provider not implement IConverterDriver")
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
