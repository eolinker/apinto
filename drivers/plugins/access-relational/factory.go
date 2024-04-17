package access_relational

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

const (
	Name = "access_relational"
)

var (
	customerVar eosc.ICustomerVar
	once        sync.Once
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return nil
}

func Create(id string, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	once.Do(func() {
		bean.Autowired(&customerVar)
	})
	err := Check(v, workers)
	if err != nil {
		return nil, err
	}

	ar := &AccessRelational{}

	iResponse, handlers := ar.parseConfig(v)
	ar.response = iResponse
	ar.rules = handlers
	return ar, nil
}
