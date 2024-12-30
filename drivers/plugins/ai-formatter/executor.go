package ai_formatter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type executor struct {
	drivers.WorkerBase
	provider string
	model    string
	extender map[string]interface{}
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	//convert.SetAIProvider(ctx, e.provider)
	//convert.SetAIModel(ctx, e.model)
	//v, has := convertManger.Get(e.provider)
	//if !has {
	//	return errors.New("provider not implement IConverterDriver")
	//}
	//converter, has := v.GetConverter(e.model)
	//if !has {
	//	return errors.New("invalid model")
	//}
	//err := converter.RequestConvert(ctx, e.extender)
	//if err != nil {
	//	return err
	//}
	//if next != nil {
	//	err = next.DoChain(ctx)
	//	if err != nil {
	//		return err
	//	}
	//}
	//return converter.ResponseConvert(ctx)
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

func (e *executor) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	//v, has := convert.Get(string(cfg.Provider))
	//if !has {
	//	return errors.New("provider not implement IConverterDriver")
	//}
	//_, has = v.GetConverter(cfg.Model)
	//if !has {
	//	return errors.New("invalid model")
	//}
	//f, has := v.GetModel(cfg.Model)
	//if !has {
	//	return errors.New("invalid model")
	//}

	//extender, err := f(cfg.Config)
	//if err != nil {
	//	return err
	//}
	//e.provider = string(cfg.Provider)
	//e.model = cfg.Model
	//e.extender = extender
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
