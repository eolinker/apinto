package app

import (
	"errors"
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
	
	"github.com/eolinker/eosc"
)

var (
	errorConfigType = errors.New("error config type")
)

//driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type driver struct {
	profession string
	driver     string
	label      string
	desc       string
	configType reflect.Type
}

//ConfigType 返回service_http驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *driver) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

//Create 创建service_http驱动的实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	
	return &app{
		id:         id,
		config:     cfg,
		appManager: appManager,
	}, nil
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, errorConfigType
	}
	for _, a := range conf.Auth {
		err := application.CheckPosition(a.Position)
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}
