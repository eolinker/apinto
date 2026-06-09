package access_relation_redis

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "access_relation_redis"
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

	ar := &AccessRelationRedis{}
	err = ar.reset(v)
	if err != nil {
		return nil, err
	}
	return ar, nil
}
