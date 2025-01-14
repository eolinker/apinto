package params_check_v2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "params_check_v2"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg := (*Param)(conf)

	err := checkParam(cfg)
	if err != nil {
		return nil, err
	}
	ck, err := newParamChecker(cfg)
	if err != nil {
		return nil, err
	}

	return &executor{
		WorkerBase: drivers.Worker(id, name),
		ck:         ck,
		logic:      cfg.Logic,
	}, nil
}
