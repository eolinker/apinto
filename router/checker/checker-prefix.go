package checker

import (
	"fmt"
	"strings"
)

type PrefixChecker struct {
	prefix string
}

func (p *PrefixChecker) Key() string {
	return fmt.Sprintf("^=%s",p.prefix)
}

func newCheckerPrefix(prefix string) *PrefixChecker {
	return &PrefixChecker{prefix: prefix}
}

func (p *PrefixChecker) Value() string {
	return p.prefix
}

func (p *PrefixChecker) Check(v string,has bool) bool{
	if !has{
		return false
	}
	return strings.HasPrefix(v,p.prefix)
}

func (p *PrefixChecker) CheckType() CheckType {
	return CheckTypePrefix
}