package checker

import (
	"fmt"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

//SubChecker 实现了Checker接口，能进行子串匹配
type SubChecker struct {
	sub string
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (p *SubChecker) Key() string {
	return fmt.Sprintf("=*%s*", p.sub)
}

//Value 返回路由指标检查器的检测值
func (p *SubChecker) Value() string {
	return p.sub
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (p *SubChecker) Check(v string, has bool) bool {
	//当待检测的路由指标值存在 且 检查器的检测值为其子串时匹配成功
	if !has {
		return false
	}
	return strings.Contains(v, p.sub)
}

//CheckType 返回检查器的类型值
func (p *SubChecker) CheckType() http_service.CheckType {
	return http_service.CheckTypeSub
}

//newCheckerAll 创建一个子串匹配类型的检查器
func newCheckerSub(sub string) *SubChecker {
	return &SubChecker{sub: sub}
}
