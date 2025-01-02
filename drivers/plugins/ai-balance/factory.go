package ai_balance

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "ai_balance"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
	eosc.IExtenderDriverFactory
}

func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create),
	}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	_, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = w.reset()
	if err != nil {
		return nil, err
	}
	return w, err
}
