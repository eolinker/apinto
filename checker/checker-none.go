package checker

import (
	"strings"

	"github.com/eolinker/eosc/http"
)

var (
	globalCheckerNone = &checkerNone{}
)

//checkerAll 实现了Checker接口，能进行空值匹配
type checkerNone struct {
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (t *checkerNone) Key() string {
	return "$"
}

//Value 返回路由指标检查器的检测值
func (t *checkerNone) Value() string {
	return "$"
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (t *checkerNone) Check(v string, has bool) bool {
	//当待检测的路由指标值存在且值为空时匹配成功
	if !has {
		return false
	}
	return strings.TrimSpace(v) == ""
}

//CheckType 返回检查器的类型值
func (t *checkerNone) CheckType() http.CheckType {
	return http.CheckTypeNone
}

//newCheckerAll 创建一个空值匹配类型的检查器
func newCheckerNone() *checkerNone {
	return globalCheckerNone
}
