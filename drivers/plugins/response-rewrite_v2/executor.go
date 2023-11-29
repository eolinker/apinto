package response_rewrite_v2

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers"
)

var _ eocontext.IFilter = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	handlers []*responseRewrite
}

type responseRewrite struct {
	matcher        *matcher
	rewriteHandler *rewriteHandler
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if next != nil {
		err = next.DoChain(ctx)
	}
	for _, handler := range e.handlers {
		variables, ok := handler.matcher.Match(ctx)
		if ok {
			handler.rewriteHandler.Rewrite(ctx, variables)
			break
		}
	}
	return
}

func (e *executor) Destroy() {
	e.handlers = nil
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
