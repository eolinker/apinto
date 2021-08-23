package checker

import "fmt"

//checkerAll 实现了Checker接口，能进行全等匹配
type checkerEqual struct {
	value string
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (e *checkerEqual) Key() string {
	return fmt.Sprintf("=%s",e.value)
}

//Value 返回路由指标检查器的检测值
func (e *checkerEqual) Value() string {
	return e.value
}

//newCheckerAll 创建一个全等匹配类型的检查器
func newCheckerEqual(value string) *checkerEqual {
	return &checkerEqual{value: value}
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (e *checkerEqual) Check(v string,has bool) bool{
	//当待检测路由指标值存在  且值与检查器的检测值相等时匹配成功
	if !has{
		return false
	}
	return v == e.value
}

//CheckType 返回检查器的类型值
func (e *checkerEqual) CheckType() CheckType {
	return CheckTypeEqual
}
