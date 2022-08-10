package nsq

import (
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func Check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, errConfigType
	}

	nsqConf := conf
	if nsqConf == nil {
		return nil, errNsqConfNull
	}
	if nsqConf.Topic == "" {
		return nil, errTopicNull
	}
	if len(nsqConf.Address) == 0 {
		return nil, errAddressNull
	}
	if nsqConf.Type == "" {
		nsqConf.Type = "line"
	}
	switch nsqConf.Type {
	case "line", "json":
	default:
		return nil, errFormatterType
	}

	if len(nsqConf.Formatter) == 0 {
		return nil, errFormatterConf
	}

	return nsqConf, nil
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	conf, err := Check(v)
	if err != nil {
		return nil, err
	}
	worker := &NsqOutput{
		id: id,
	}
	worker.config = conf
	return worker, nil

}
