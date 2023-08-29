package separator

import (
	"fmt"
	"net/url"
	"strconv"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ ICounter = (*FormDataCounter)(nil)

type FormDataCounter struct {
	typ      string
	split    string
	name     string
	max      int64
	splitLen int
}

func (f *FormDataCounter) Name() string {
	return f.name
}

func NewFormDataCounter(rule *CountRule) (*FormDataCounter, error) {
	var splitLen int
	if rule.SeparatorType == LengthCountType {
		var err error
		splitLen, err = strconv.Atoi(rule.Separator)
		if err != nil {
			splitLen = 1000
		}
	}
	return &FormDataCounter{name: rule.Key, split: rule.Separator, typ: rule.SeparatorType, max: rule.Max, splitLen: splitLen}, nil
}

func (f *FormDataCounter) Count(ctx http_service.IHttpContext) (int64, error) {
	body, _ := ctx.Request().Body().RawBody()
	u, err := url.ParseQuery(string(body))
	if err != nil {
		return -1, fmt.Errorf("parse form data error:%v", err)
	}
	switch f.typ {
	case SplitCountType:
		return splitCount(u.Get(f.name), f.split), nil
	case LengthCountType:
		value := u.Get(f.name)
		l := len([]rune(value))
		if l%f.splitLen == 0 {
			return int64(l / f.splitLen), nil
		}
		return int64(l/f.splitLen + 1), nil
	}
	return splitCount(u.Get(f.name), f.split), nil
}

func (f *FormDataCounter) Max() int64 {
	return f.max
}
