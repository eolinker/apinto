// Package pricing_driver 提供价格计算器 worker 的注册入口。
//
// 该 worker 在 Reset 时按 provider/resource/phase 三元组将自身注册到
// 顶层 pricing 包的全局表，供 drivers/plugins/billing 在拦截阶段查找并执行计费。
// 注：包名为 pricing_driver 而非 pricing，避免与顶层契约包 pricing 冲突。
package pricing_driver

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// Name 驱动名称，对应配置 profession=resource、driver=pricing。
const Name = "pricing"

// Register 注册价格计算器驱动。
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

// NewFactory 创建价格计算器驱动工厂。
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
