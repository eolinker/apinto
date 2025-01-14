package params_check_v2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

const (
	Name = "params_check_v2"
)

func Register(register eosc.IExtenderDriverRegister) {
	log.Debug("register params_check_v2 is ", Name)
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := checkParam(conf)
	if err != nil {
		return nil, err
	}
	cks := make([]IParamChecker, 0, len(conf.Params))
	for _, p := range conf.Params {
		ck, err := newParamChecker(p)
		if err != nil {
			return nil, err
		}
		cks = append(cks, ck)
	}

	return &executor{
		WorkerBase: drivers.Worker(id, name),
		cks:        cks,
		logic:      conf.Logic,
	}, nil
}
