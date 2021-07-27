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
		return parseValue(p)
	}

	tp := strings.TrimSpace(pattern[:i])
	v := strings.TrimSpace(pattern[i+1:])

	switch tp {
	case "^":
		if len(v) > 0 {
			if v[0] == '*' {
				return newCheckerSuffix(v[1:]), nil
			}
		}
		return newCheckerPrefix(v), nil
	case "":
		return parseValue(v)
	case "!":
		return newCheckerNotEqual(v), nil
	case "~":
		return newCheckerRegexp(v)
	case "~*":
		return newCheckerRegexpG(v)
	}

	return nil, fmt.Errorf("%s:%w", pattern, ErrorUnknownExpression)
}

func parseValue(v string)(Checker,error)  {
	switch v {
	case "*":
		return newCheckerAll(), nil
	case "**":
		return newCheckerExist(), nil
	case "!":
		return newCheckerNotExits(), nil
	case "$":
		return newCheckerNone(), nil
	default:
		if len(v) == 0 {
			return newCheckerAll(), nil
		}
		l:=len(v)
		if len(v)>1 && v[0]=='*' && v[l-1]!= '*' {
			return  newCheckerSuffix(v[1:]),nil
		}
		if len(v)>1 && v[l-1]=='*' && v[0]!= '*'{
			return newCheckerPrefix(v[:l-1]),nil
		}
		if len(v)>2 && v[0] == '*' && v[l-1] == '*'{
			return newCheckerSub(v[1:l-1]),nil
		}
		return newCheckerEqual(v), nil
	}
}