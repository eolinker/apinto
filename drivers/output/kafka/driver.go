package kafka

import (
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	_, err := check(v)
	return err
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func check(v interface{}) (*ProducerConfig, error) {
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

<<<<<<< ours
func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
=======
func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	cfg, err := check(v)
	if err != nil {
		return nil, err
	}

>>>>>>> theirs
	worker := &Output{
		id:       id,
		name:     name,
		producer: nil,
		config:   cfg,
	}

	return worker, err
}
