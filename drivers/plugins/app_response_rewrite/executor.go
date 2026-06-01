package app_response_rewrite

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/response"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type Executor struct {
	drivers.WorkerBase
	response response.IResponse
}

func (a *Executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	// 判断是否是websocket
	return http_service.DoHttpFilter(a, ctx, next)
}

func (a *Executor) Destroy() {
}

func (a *Executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {

	if next != nil {
		err := next.DoChain(ctx)
		if err != nil {
			if ctx.GetLabel("auth_status") == "fail" && a.response != nil {
				a.response.Response(ctx)
			}
			return err
		}
	}

	return nil
}

func (a *Executor) DoWebsocketFilter(ctx http_service.IWebsocketContext, next eocontext.IChain) error {

	if next != nil {
		err := next.DoChain(ctx)
		if err != nil {
			if a.response != nil {
				a.response.Response(ctx)
			}
			return err
		}
	}

	return nil
}

func (a *Executor) Start() error {
	return nil
}

func (a *Executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (a *Executor) Stop() error {
	return nil
}

func (a *Executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
