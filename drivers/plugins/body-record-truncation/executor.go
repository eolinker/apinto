package body_record_truncation

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers"
)

var ()

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	bodySize int64
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if ctx.Request().Method() == "POST" || ctx.Request().Method() == "PUT" || ctx.Request().Method() == "PATCH" {
		if e.bodySize != 0 && int64(ctx.Request().ContentLength()) > e.bodySize {
			// 当请求体大小大于限制时，截断请求体
			entry := ctx.GetEntry()
			body := entry.Read("ctx_request_body")
			v, _ := body.(string)
			ctx.SetLabel("request_body", v[:e.bodySize])
			ctx.WithValue("request_body_complete", 0)
		} else {
			ctx.WithValue("request_body_complete", 1)
		}
	}
	if next != nil {
		err = next.DoChain(ctx)
	}
	if e.bodySize != 0 && int64(ctx.Response().ContentLength()) > e.bodySize {
		// 当响应体大小大于限制时，截断响应体
		entry := ctx.GetEntry()
		body := entry.Read("ctx_response_body")
		v, _ := body.(string)
		ctx.SetLabel("response_body", v[:e.bodySize])
		ctx.WithValue("response_body_complete", 0)
	} else {
		ctx.WithValue("response_body_complete", 1)
	}
	return err
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
