package checker

import (
	"fmt"
)

// checkerAll 实现了Checker接口，能进行非等匹配
type checkerNotEqual struct {
	value string
}

// Key 返回路由指标检查器带有完整规则符号的检测值
func (e *checkerNotEqual) Key() string {
	return fmt.Sprintf("!=%s", e.value)
}

// Value 返回路由指标检查器的检测值
func (e *checkerNotEqual) Value() string {
	return e.value
}

// Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (e *checkerNotEqual) Check(v string, has bool) bool {
	//当待检测路由指标值存在且与检查器的检测值不相等时匹配成功
	if !has {
		return false
	}
	return v != e.value
}

// CheckType 返回检查器的类型值
func (e *checkerNotEqual) CheckType() CheckType {
	return CheckTypeNotEqual
}

// newCheckerAll 创建一个非等匹配类型的检查器
func newCheckerNotEqual(value string) *checkerNotEqual {
	return &checkerNotEqual{value: value}
}
