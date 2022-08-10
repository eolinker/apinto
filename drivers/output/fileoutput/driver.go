package fileoutput

import (
	"reflect"

	"github.com/eolinker/eosc"
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
		return nil, errorConfigType
	}

	fileConf := conf
	if fileConf == nil {
		return nil, errorNilConfig
	}

	if fileConf.Dir == "" {
		return nil, errorConfDir
	}
	if fileConf.File == "" {
		return nil, errorConfFile
	}
	if fileConf.Period != "day" && fileConf.Period != "hour" {
		return nil, errorConfPeriod
	}

	if fileConf.Expire == 0 {
		fileConf.Expire = 3
	}
	if fileConf.Type == "" {
		fileConf.Type = "line"
	}

	if len(fileConf.Formatter) == 0 {
		return nil, errFormatterConf
	}
	return conf, nil
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {

	cfg, err := Check(v)
	if err != nil {
		return nil, err
	}
	worker := &FileOutput{
		id:     id,
		name:   name,
		config: cfg,
	}
	return worker, err
}
