package access_hierarchy

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

const (
	// Name 定义驱动的唯一标识名
	Name = "access_hierarchy"
)

var (
	// customerVar 自动注入的网关自定义变量模块，用于读取运行期租户、资源组绑定关系
	customerVar eosc.ICustomerVar
	once        sync.Once
)

func init() {
	once.Do(func() {
		// 自动注入 eosc 容器内的 ICustomerVar 单例
		bean.Autowired(&customerVar)
	})
}

// Register 用于向系统驱动管理器注册本插件的工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

// NewFactory 构造工厂实例，指定其 Config 类型、Create 与 Check 检查函数
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}

// Check 执行前置参数静态配置校验，本插件无强校验规则
func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

// Create 在控制台新建、或者网关重启/热更新插件实例时被调用
func Create(id string, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := Check(v, workers)
	if err != nil {
		return nil, err
	}

	ah := &AccessHierarchy{}
	// 解析控制台下发的 YAML/JSON 规则及 Response 配置
	iResponse, handlers := ah.parseConfig(v)
	ah.response = iResponse
	ah.rules = handlers
	return ah, nil
}
