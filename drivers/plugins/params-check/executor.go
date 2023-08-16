package params_check

import (
	"errors"
	"fmt"
	"mime"
	"net/url"
	"reflect"
	"strings"

	"github.com/ohler55/ojg/jp"

	"github.com/ohler55/ojg/oj"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
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
	return
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func bodyCheck(ctx http_service.IHttpContext, checkers []*paramChecker) error {
	if len(checkers) < 1 {
		return nil
	}
	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
	var body interface{}
	var bodyCheckerFunc func(interface{}, *paramChecker) error
	switch contentType {
	case MultipartForm, FormData:
		bodyCheckerFunc = formChecker
		body, _ = ctx.Request().Body().BodyForm()
	case JSON:
		bodyCheckerFunc = jsonChecker
		data, _ := ctx.Request().Body().RawBody()
		if string(data) == "" {
			data = []byte("{}")
		}
		o, err := oj.Parse(data)
		if err != nil {
			return fmt.Errorf("parse json error: %v,body is %s", err, string(data))
		}
		body = o
	}

	for _, c := range checkers {
		if reflect.ValueOf(bodyCheckerFunc).IsNil() {
			continue
		}
		err := bodyCheckerFunc(body, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func formChecker(body interface{}, checker *paramChecker) error {
	params, ok := body.(url.Values)
	if !ok {
		return fmt.Errorf("error body type")
	}
	v := params.Get(checker.name)
	match := checker.Check(v, len(v) > 0)
	if !match {
		return fmt.Errorf(errParamCheck, "body", checker.name)
	}
	return nil
}

func jsonChecker(body interface{}, checker *paramChecker) error {
	name := checker.name
	if !strings.HasPrefix(name, "$.") {
		name = "$." + name
	}
	x, err := jp.ParseString(name)
	if err != nil {
		return err
	}
	result := x.Get(body)
	if len(result) < 1 {
		match := checker.Check("", false)
		if !match {
			return fmt.Errorf(errParamCheck, "body", checker.name, checker.name)
		}
	}
	v, ok := result[0].(string)
	if !ok {
		return fmt.Errorf("error body param type")
	}
	match := checker.Check(v, len(v) > 0)
	if !match {
		return fmt.Errorf(errParamCheck, "body", checker.name, checker.name)
	}
	return nil
}
