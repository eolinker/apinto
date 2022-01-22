package kafka

import (
	"github.com/eolinker/eosc"
	"reflect"
	"sync"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) check(v interface{}) (*ProducerConfig, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	pConf, err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	return pConf, nil

}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	worker := &Output{
		Driver: d,
		id:     id,
		enable: false,
		locker: &sync.Mutex{},
	}
	err := worker.Reset(v, workers)
	return worker, err
}
