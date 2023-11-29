package params_check

import (
	"errors"
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/drivers"
)

var (
	MultipartForm = "multipart/form-data"
	FormData      = "application/x-www-form-urlencoded"
	JSON          = "application/json"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	headerChecker []*paramChecker
	queryChecker  []*paramChecker
	bodyChecker   []*paramChecker
}

type paramChecker struct {
	name string
	checker.Checker
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

var errParamCheck = "Can not find the %s param \"%s\" or the \"%s\" is illegal"

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	for _, c := range e.headerChecker {
		v := ctx.Request().Header().GetHeader(c.name)
		match := c.Check(v, len(v) > 0)
		if !match {
			errInfo := fmt.Sprintf(errParamCheck, "header", c.name, c.name)
			ctx.Response().SetStatus(401, "401")
			ctx.Response().SetBody([]byte(errInfo))
			return errors.New(errInfo)
		}
	}
	for _, c := range e.queryChecker {
		v := ctx.Request().URI().GetQuery(c.name)
		match := c.Check(v, len(v) > 0)
		if !match {
			errInfo := fmt.Sprintf(errParamCheck, "query", c.name, c.name)
			ctx.Response().SetStatus(401, "401")
			ctx.Response().SetBody([]byte(errInfo))
			return errors.New(errInfo)
		}
	}
	err = bodyCheck(ctx, e.bodyChecker)
	if err != nil {
		ctx.Response().SetStatus(401, "401")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (e *executor) Destroy() {
	e.headerChecker = nil
	e.queryChecker = nil
	e.bodyChecker = nil
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
