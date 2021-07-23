package checker

import (
	"errors"
	"fmt"
	"strings"
)

var(
	ErrorUnknownExpression = errors.New("unknown expression")
)
type Checker interface {
	Check(v string,has bool) bool
	Key()string
	CheckType() CheckType
	Value()string
}

func Parse(pattern string)(Checker,error)  {
	i:=strings.Index(pattern,"=")

	if i < 0{
		p:=strings.TrimSpace(pattern)
		switch p {
		case "*":
			return newCheckerAll(),nil
		case "**":
			return newCheckerExist(),nil
		case "!":
			return newCheckerNotExits(),nil
		case "$":
			return newCheckerNone(),nil
		default:
			if len(p) == 0{
				return newCheckerAll(),nil
			}
			return newCheckerEqual(p),nil
		}
	}

	tp:= strings.TrimSpace(pattern[:i])
	v:= strings.TrimSpace(pattern[i+1:])

	switch tp{
	case "^":
		return newCheckerPrefix(v),nil
	case "":
		if len(v) == 0{
			return newCheckerAll(),nil
		}
		return newCheckerEqual(v),nil
	case "!":
		return newCheckerNotEqual(v),nil
	case "~":
		return newCheckerRegexp(v)
	case "~*":
		return newCheckerRegexpG(v)
	}

	return nil,fmt.Errorf("%s:%w",pattern,ErrorUnknownExpression)
}

