package checker

import (
	"fmt"
	"strings"
)

// SuffixChecker 实现了Checker接口，能进行后缀匹配
type SuffixChecker struct {
	suffix string
}

// Key 返回路由指标检查器带有完整规则符号的检测值
func (p *SuffixChecker) Key() string {
	return fmt.Sprintf("=*%s", p.suffix)
}

// Value 返回路由指标检查器的检测值
func (p *SuffixChecker) Value() string {
	return p.suffix
}

// Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (p *SuffixChecker) Check(v string, has bool) bool {
	//当待检测的路由指标值存在 且 检查器的检测值为其后缀时匹配成功
	if !has {
		return false
	}
	return strings.HasSuffix(v, p.suffix)
}

// CheckType 返回检查器的类型值
func (p *SuffixChecker) CheckType() CheckType {
	return CheckTypeSuffix
}

// newCheckerAll 创建一个后缀匹配类型的检查器
func newCheckerSuffix(suffix string) *SuffixChecker {
	return &SuffixChecker{suffix: suffix}
}
