package separator

import (
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	ArrayCountType  = "array"
	SplitCountType  = "splite"
	LengthCountType = "length"
	CountTypes      = []string{
		ArrayCountType,
		SplitCountType,
		LengthCountType,
	}
)

type CountRule struct {
	RequestBodyType string `json:"request_body_type" label:"请求体类型" enum:"form-data,json"`
	Key             string `json:"key" label:"参数名称（支持json path）"`
	Separator       string `json:"separator" label:"分隔符" switch:"separator_type===splite"`
	SeparatorType   string `json:"separator_type" label:"分割类型" enum:"splite,array,length"`
	Max             int64  `json:"max" label:"计数最大值"`
}

type ICounter interface {
	Count(ctx http_service.IHttpContext) (int64, error)
	Max() int64
	Name() string
}

func GetCounter(rule *CountRule) (ICounter, error) {
	if rule == nil && rule.Key == "" {
		return NewEmptyCounter(), nil
	}
	switch strings.ToLower(rule.RequestBodyType) {
	case "form-data":
		return NewFormDataCounter(rule)
	case "multipart-formdata":
		return NewFileCounter(rule)
	case "json":
		return NewJsonCounter(rule)
	default:
		return NewEmptyCounter(), nil
	}
}

func splitCount(origin string, split string) int64 {
	if len(split) == 0 {
		return 0
	}

	vs := strings.Split(origin, string(split[0]))
	var count int64 = 0
	for _, v := range vs {
		if v != "" {
			childCount := splitCount(v, split[1:])
			if childCount == 0 {
				count += 1
			} else {
				count += childCount
			}
		}

	}

	return count
}
