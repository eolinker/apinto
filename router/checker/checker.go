package checker

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrorUnknownExpression = errors.New("unknown expression")
)

type Checker interface {
	Check(v string, has bool) bool
	Key() string
	CheckType() CheckType
	Value() string
}

func Parse(pattern string) (Checker, error) {
	i := strings.Index(pattern, "=")

	if i < 0 {
		p := strings.TrimSpace(pattern)
		switch p {
		case "*": //任意
			return newCheckerAll(), nil
		case "**": //只要key存在
			return newCheckerExist(), nil
		case "!": //key不存在时
			return newCheckerNotExits(), nil
		case "$": //空值
			return newCheckerNone(), nil
		default:
			if len(p) == 0 {
				return newCheckerAll(), nil
			}
			// 测全等
			return newCheckerEqual(p), nil
		}
	}

	tp := strings.TrimSpace(pattern[:i])
	v := strings.TrimSpace(pattern[i+1:])

	switch tp {
	case "^": //前缀匹配
		return newCheckerPrefix(v), nil
	case "": //全等
		if len(v) == 0 {
			return newCheckerAll(), nil
		}
		return newCheckerEqual(v), nil
	case "!": //不等
		return newCheckerNotEqual(v), nil
	case "~": //区分大小写的正则
		return newCheckerRegexp(v)
	case "~*": //不区分大小写的正则
		return newCheckerRegexpG(v)
	}

	return nil, fmt.Errorf("%s:%w", pattern, ErrorUnknownExpression)
}
