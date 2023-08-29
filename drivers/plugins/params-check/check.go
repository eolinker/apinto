package params_check

import (
	"fmt"
	"mime"
	"net/url"
	"reflect"
	"strings"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

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
