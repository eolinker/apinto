package checker

import (
	"errors"
	"fmt"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

var (
	errorUnknownExpression = errors.New("unknown expression")
)

//Parse 可根据路由指标字符串生成相应的检查器
func Parse(pattern string) (http_service.Checker, error) {
	i := strings.Index(pattern, "=")

	if i < 0 {
		p := strings.TrimSpace(pattern)
		return parseValue(p)
	}

	tp := strings.TrimSpace(pattern[:i])
	v := strings.TrimSpace(pattern[i+1:])

	switch tp {
	case "^":
		if len(v) > 0 {
			if v[0] == '*' {
				return newCheckerSuffix(v[1:]), nil // ^= *abc 后缀匹配
			}
		}
		return newCheckerPrefix(v), nil // ^= abc 前缀匹配
	case "":
		return parseValue(v)
	case "!":
		return newCheckerNotEqual(v), nil //!= 不等于
	case "~":
		return newCheckerRegexp(v) //~= 区分大小写的正则
	case "~*":
		return newCheckerRegexpG(v) //~*=  不区分大小写的正则
	}

	return nil, fmt.Errorf("%s:%w", pattern, errorUnknownExpression)
}

//parseValue 根据不带等号的指标字符串生成检查器
func parseValue(v string) (http_service.Checker, error) {
	switch v {
	case "*": //任意
		return newCheckerAll(), nil
	case "**": //只要key存在
		return newCheckerExist(), nil
	case "!": //key不存在时
		return newCheckerNotExits(), nil
	case "$": //空值
		return newCheckerNone(), nil
	default:
		if len(v) == 0 {
			return newCheckerAll(), nil //任意
		}
		l := len(v)
		if len(v) > 1 && v[0] == '*' && v[l-1] != '*' {
			return newCheckerSuffix(v[1:]), nil //*.abc.com 后缀匹配
		}
		if len(v) > 1 && v[l-1] == '*' && v[0] != '*' {
			return newCheckerPrefix(v[:l-1]), nil //abc*前缀匹配
		}
		if len(v) > 2 && v[0] == '*' && v[l-1] == '*' {
			return newCheckerSub(v[1 : l-1]), nil //*abc*子串匹配
		}
		return newCheckerEqual(v), nil //全等
	}
}
