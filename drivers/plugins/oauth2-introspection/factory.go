package oauth2_introspection

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/app/manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

const (
	Name = "oauth2-introspection"
)

var (
	ones       sync.Once
	appManager manager.IManager
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		bean.Autowired(&appManager)
	})
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := Check(conf)
	if err != nil {
		return nil, err
	}
	e := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = e.reset(conf)
	if err != nil {
		return nil, err
	}
	return e, nil
}
