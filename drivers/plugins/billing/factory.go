// Package billing 提供通用计费拦截插件，覆盖 AI 大模型与标准 API 两类调用场景。
//
// 插件职责：
//   - 在 HttpFilter 链中拦截请求/响应
//   - 按 provider/resource/phase 三元组查找已注册的价格计算器
//   - 完成两阶段余额扣减（PreDeduct / Settle / Rollback）
//   - 防异步任务重复计费（Redis SetNX + 本地内存双层）
//   - 将计费结果写入 ctx label，供日志/审计下游消费
package billing

import (
	"sync"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// Name 插件名称，对应 profession=plugin、driver=billing。
const Name = "billing"

var (
	workerResources eosc.IWorkers
	once            sync.Once
)

// Register 注册 billing 插件驱动。
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

// Factory 插件工厂，组合 drivers.NewFactory 实现并附带一次性的 bean 自动装配。
type Factory struct {
	eosc.IExtenderDriverFactory
}

// NewFactory 创建 billing 插件驱动工厂。
func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create),
	}
}

// Create 拦截 IExtenderDriverFactory.Create 调用，确保 workerResources 已自动装配后再创建驱动。
func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	once.Do(func() {
		bean.Autowired(&workerResources)
	})
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}
