package checker

import (
	"fmt"
	"github.com/eolinker/eosc/log"
	"strings"
)

//PrefixChecker 实现了Checker接口，能进行前缀匹配
type PrefixChecker struct {
	prefix string
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (p *PrefixChecker) Key() string {
	return fmt.Sprintf("%s*", p.prefix)
}

//Value 返回路由指标检查器的检测值
func (p *PrefixChecker) Value() string {
	return p.prefix
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (p *PrefixChecker) Check(v string, has bool) bool {
	log.Debug("PrefixChecker:check", v, has, " on ", p.prefix)
	//当待检测的路由指标值存在 且 检查器的检测值为其前缀时匹配成功
	if !has {
		return false
	}

	return strings.HasPrefix(v, p.prefix)
}

//CheckType 返回检查器的类型值
func (p *PrefixChecker) CheckType() CheckType {
	return CheckTypePrefix
}

//newCheckerAll 创建一个前缀匹配类型的检查器
func newCheckerPrefix(prefix string) *PrefixChecker {
	return &PrefixChecker{prefix: prefix}
}
