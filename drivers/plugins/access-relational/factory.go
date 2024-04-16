package access_relational

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

const (
	Name = "access_relational"
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
	err := Check(v, workers)
	if err != nil {
		return nil, err
	}
	iResponse, handlers := parseConfig(v)
	ar := &AccessRelational{
		response: iResponse,
		rules:    handlers,
	}
	bean.Autowired(&ar.data)
	return ar, nil
}
