package checker

var (
	globalCheckerExist    = &checkerExist{}
	globalCheckerNotExist = &checkerNotExits{}
)

// checkerAll 实现了Checker接口，能进行存在匹配
type checkerExist struct {
}

// Key 返回路由指标检查器带有完整规则符号的检测值
func (t *checkerExist) Key() string {
	return "**"
}

// Value 返回路由指标检查器的检测值
func (t *checkerExist) Value() string {
	return "**"
}

// Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (t *checkerExist) Check(v string, has bool) bool {
	//当待检测的路由指标值存在且长度大于0时匹配成功
	return has && len(v) > 0
}

// CheckType 返回检查器的类型值
func (t *checkerExist) CheckType() CheckType {
	return CheckTypeExist
}

// newCheckerAll 创建一个存在匹配类型的检查器
func newCheckerExist() *checkerExist {
	return globalCheckerExist
}

// checkerAll 实现了Checker接口，能进行不存在匹配
type checkerNotExits struct {
}

// Key 返回路由指标检查器带有完整规则符号的检测值
func (c *checkerNotExits) Key() string {
	return "!"
}

// Value 返回路由指标检查器的检测值
func (c *checkerNotExits) Value() string {
	return "!"
}

// Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (c *checkerNotExits) Check(v string, has bool) bool {
	//当待检测的路由指标值不存在时匹配成功
	return !has
}

// CheckType 返回检查器的类型值
func (c *checkerNotExits) CheckType() CheckType {
	return CheckTypeNotExist
}

// newCheckerAll 创建一个不存在匹配类型的检查器
func newCheckerNotExits() *checkerNotExits {
	return globalCheckerNotExist
}
