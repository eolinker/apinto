package params_check_v2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/oj"
)

var (
	MultipartForm = "multipart/form-data"
	FormData      = "application/x-www-form-urlencoded"
	JSON          = "application/json"
	logicAnd      = "and"
	logicOr       = "or"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	logic string
	ck    IParamChecker
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

var errParamCheck = "Can not find the %s param \"%s\" or the \"%s\" is illegal"

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if e.ck != nil {
		headerReader := ctx.Request().Header()
		queryReader := ctx.Request().URI()
		var body interface{}
		var fn bodyCheckerFunc
		contentType := ctx.Request().Body().ContentType()
		switch contentType {
		case MultipartForm, FormData:
			body, _ = ctx.Request().Body().BodyForm()
			fn = formChecker
		case JSON:
			data, _ := ctx.Request().Body().RawBody()
			body, err = oj.Parse(data)
			if err != nil {
				ctx.Response().SetBody([]byte("parse json error: " + err.Error()))
				ctx.Response().SetStatus(400, "Bad Request")
				return err
			}
			fn = jsonChecker

		}
		ok := e.ck.Check(e.logic, headerReader, queryReader, body, fn)
		if !ok {
			ctx.Response().SetStatus(401, "401")
			ctx.Response().SetBody([]byte("param check failed"))
			return err
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (e *executor) Destroy() {
	e.ck = nil
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
