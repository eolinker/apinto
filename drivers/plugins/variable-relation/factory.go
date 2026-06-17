package variable_relation

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

const (
	Name = "variable_relation"
)

var (
	customerVar eosc.ICustomerVar
	once        sync.Once
)

func init() {
	once.Do(func() {
		bean.Autowired(&customerVar)
	})
}

// Register 注册插件
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

// NewFactory 插件工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}

// Check 检查配置
func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

// Create 创建插件实例
func Create(id string, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := Check(v, workers)
	if err != nil {
		return nil, err
	}

	vr := &VariableRelation{}
	err = vr.parseConfig(v)
	if err != nil {
		return nil, err
	}
	return vr, nil
}
