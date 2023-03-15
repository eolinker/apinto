package http_to_dubbo2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"time"
)

var _ eocontext.IFilter = (*ToDubbo2)(nil)
var _ http_context.HttpFilter = (*ToDubbo2)(nil)

type ToDubbo2 struct {
	drivers.WorkerBase
	service string
	method  string
	params  []param
}

func (p *ToDubbo2) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {

	retryValue := ctx.Value(eocontext.CtxKeyRetry)
	retry, ok := retryValue.(int)
	if !ok {
		retry = 1
	}

	timeoutValue := ctx.Value(eocontext.CtxKeyTimeout)
	timeout, ok := timeoutValue.(time.Duration)
	if !ok {
		timeout = 3000 * time.Millisecond
	}

	complete := NewComplete(retry, timeout, p.service, p.method, p.params)
	ctx.SetCompleteHandler(complete)

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (p *ToDubbo2) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(p, ctx, next)
}

type param struct {
	className string
	fieldName string
}

func (p *ToDubbo2) Start() error {
	return nil
}

func (p *ToDubbo2) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	p.service = conf.Service
	p.method = conf.Method

	params := make([]param, 0, len(conf.Params))

	for _, val := range conf.Params {
		params = append(params, param{
			className: val.ClassName,
			fieldName: val.FieldName,
		})
	}
	p.params = params
	return nil
}

func (p *ToDubbo2) Stop() error {
	return nil
}

func (p *ToDubbo2) Destroy() {
}

func (p *ToDubbo2) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
