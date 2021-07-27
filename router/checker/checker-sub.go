package checker

import (
	"fmt"
	"strings"
)

type SubChecker struct {
	sub string
}

func (p *SubChecker) Key() string {
	return fmt.Sprintf("=*%s*",p.sub)
}

func newCheckerSub(sub string) *SubChecker {
	return &SubChecker{sub: sub}
}

func (p *SubChecker) Value() string {
	return p.sub
}

func (p *SubChecker) Check(v string,has bool) bool{
	if !has{
		return false
	}
	return strings.HasPrefix(v,p.sub)
}

func (p *SubChecker) CheckType() CheckType {
	return CheckTypeSub
}