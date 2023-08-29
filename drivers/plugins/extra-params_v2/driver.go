package extra_params_v2

import (
	"reflect"
	"sync"

	"github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params/concat"

	"github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params/datetime"
	"github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params/md5"
	"github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params/timestamp"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
)

var (
	once sync.Once
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return conf.doCheck()
}

func check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	once.Do(func() {
		datetime.Register()
		md5.Register()
		timestamp.Register()
		concat.Register()
	})
	ep := &executor{
		WorkerBase:      drivers.Worker(id, name),
		baseParam:       generateBaseParam(conf.Params),
		requestBodyType: conf.RequestBodyType,
		errorType:       conf.ErrorType,
	}

	return ep, nil
}
