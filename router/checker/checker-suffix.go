package checker

import (
	"fmt"
	"strings"
)

type SuffixChecker struct {
	suffix string
}

func (p *SuffixChecker) Key() string {
	return fmt.Sprintf("=*%s", p.suffix)
}

func newCheckerSuffix(suffix string) *SuffixChecker {
	return &SuffixChecker{suffix: suffix}
}

func (p *SuffixChecker) Value() string {
	return p.suffix
}

func (p *SuffixChecker) Check(v string, has bool) bool {
	if !has {
		return false
	}
	return strings.HasSuffix(v, p.suffix)
}

func (p *SuffixChecker) CheckType() CheckType {
	return CheckTypeSuffix
}
