package replay_attack_defender

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "replay_attack_defender"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err := w.reset(conf)
	if err != nil {
		return nil, err
	}
	return w, nil
}
