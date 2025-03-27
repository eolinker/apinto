package response_filter

import (
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	bodyFilter   []jp.Expr
	headerFilter []string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	ctx.SetLabel("disable_stream", "true")
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	body := ctx.Response().GetBody()
	n, err := oj.Parse(body)
	if err != nil {
		return err
	}
	for _, filter := range e.bodyFilter {
		filter.Del(n)
	}
	body, err = oj.Marshal(n)
	ctx.Response().SetBody(body)
	for _, filter := range e.headerFilter {
		ctx.Response().DelHeader(filter)
	}
	return nil
}

func (e *executor) Destroy() {
	return
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
