package separator

import (
	"fmt"
	"strconv"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"
)

var _ ICounter = (*JsonCounter)(nil)

type JsonCounter struct {
	max      int64
	split    string
	expr     jp.Expr
	typ      string
	name     string
	splitLen int
}

func (j *JsonCounter) Name() string {
	return j.name
}

func NewJsonCounter(rule *CountRule) (*JsonCounter, error) {
	typeValid := false
	for _, t := range CountTypes {
		if t == rule.SeparatorType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return nil, fmt.Errorf("json count split type config error,now type is %s, need array or split", rule.SeparatorType)
	}
	expr, err := jp.ParseString(rule.Key)
	if err != nil {
		return nil, fmt.Errorf("json path parse error:%v", err)
	}
	if rule.Max < 1 || rule.Max > 2000 {
		rule.Max = 2000
	}
	var splitLen int
	if rule.SeparatorType == LengthCountType {
		splitLen, err = strconv.Atoi(rule.Separator)
		if err != nil {
			splitLen = 1000
		}
	}
	return &JsonCounter{max: rule.Max, split: rule.Separator, expr: expr, typ: rule.SeparatorType, name: strings.TrimPrefix(rule.Key, "$."), splitLen: splitLen}, nil
}

func (j *JsonCounter) Count(ctx http_service.IHttpContext) (int64, error) {
	body, _ := ctx.Request().Body().RawBody()
	obj, err := oj.Parse(body)
	if err != nil {
		return -1, fmt.Errorf("parse json body error:%v, body is %s", err, string(body))
	}
	results := j.expr.Get(obj)
	if len(results) == 0 {
		return -1, fmt.Errorf("json path %s get value is empty", j.name)
	}
	switch j.typ {
	case SplitCountType:
		origin, ok := results[0].(string)
		if !ok {
			return -1, fmt.Errorf("json path %s get value is not string", j.name)
		}
		return splitCount(origin, j.split), nil
	case ArrayCountType:
		switch v := results[0].(type) {
		case []interface{}:
			{
				return int64(len(v)), nil
			}
		case map[string]interface{}:
			return int64(len(v)), nil
		}
	case LengthCountType:
		origin, ok := results[0].(string)
		if !ok {
			return -1, fmt.Errorf("json path %s get value is not string", j.name)
		}
		l := len([]rune(origin))

		if l%j.splitLen == 0 {
			return int64(l / j.splitLen), nil
		}
		return int64(l/j.splitLen + 1), nil

	}
	return 1, nil
}

func (j *JsonCounter) Max() int64 {
	return j.max
}
