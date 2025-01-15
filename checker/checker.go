package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ohler55/ojg/jp"
)

var (
	errorUnknownExpression = errors.New("unknown expression")
)

// Checker 路由指标检查器接口
type Checker interface {
	Handler
	Key() string
	CheckType() CheckType
	Value() string
}

// Parse 可根据路由指标字符串生成相应的检查器
func Parse(pattern string) (Checker, error) {
	pattern = strings.TrimSpace(pattern)

	i := strings.Index(pattern, "=")

	if i < 0 {
		return parseValue(pattern)
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

// parseValue 根据不带等号的指标字符串生成检查器
func parseValue(v string) (Checker, error) {
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

var JsonArrayMatchAll = "all"
var JsonArrayMatchAny = "any"

func CheckJson(result interface{}, matchMode string, expr jp.Expr, ck Checker) bool {

	v := expr.Get(result)

	if len(v) < 1 {
		return false
	}

	return checkJsonArray(v, matchMode, ck)
}

func checkJsonArray(result []interface{}, matchMode string, ck Checker) bool {
	success := false
	for _, r := range result {
		ok := checkJsonParam(r, ck)
		if ok {
			success = true
			if matchMode == JsonArrayMatchAny {
				return true
			}
		} else {
			success = false
			if matchMode == JsonArrayMatchAll {
				return false
			}
		}
	}
	return success
}

func checkJsonParam(result interface{}, ck Checker) bool {
	switch t := result.(type) {
	case string:
		if !ck.Check(t, true) {
			return false
		}
	case bool:
		v := strconv.FormatBool(t)
		if !ck.Check(v, true) {
			return false
		}
	case int64:
		v := strconv.FormatInt(t, 10)
		if !ck.Check(v, true) {
			return false
		}
	case float64:
		v := strconv.FormatFloat(t, 'f', -1, 64)
		if !ck.Check(v, true) {
			return false
		}
	case []interface{}, map[string]interface{}:
		data, _ := json.Marshal(t)
		if !ck.Check(string(data), true) {
			return false
		}
	default:
		return false
	}
	return true
}
