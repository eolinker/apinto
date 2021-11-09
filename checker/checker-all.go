package checker

import "github.com/eolinker/eosc/http"

var globalCheckerAll = &checkerAll{}

//checkerAll 实现了Checker接口，能进行任意匹配
type checkerAll struct {
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (t *checkerAll) Key() string {
	return "*"
}

//Value 返回路由指标检查器的检测值
func (t *checkerAll) Value() string {
	return "*"
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (t *checkerAll) Check(v string, has bool) bool {
	//任意匹配能通过任何类型的路由指标值
	return true
}

//CheckType 返回检查器的类型值
func (t *checkerAll) CheckType() http.CheckType {
	return http.CheckTypeAll
}

//newCheckerAll 创建一个任意匹配类型的检查器
func newCheckerAll() *checkerAll {
	return globalCheckerAll
}
