package syslog

import (
	"fmt"
	"github.com/eolinker/eosc"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	"reflect"
)

const (
	driverName = "syslog"
)

//driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type driver struct {
	profession string
	name       string
	driver     string
	label      string
	desc       string
	configType reflect.Type
	params     map[string]string
}

//ConfigType 返回syslog驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建syslog驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, ok := v.(*DriverConfig)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(v))
	}
	c, err := toConfig(conf)
	if err != nil {
		return nil, err
	}
	a := &syslog{
		id:                 id,
		name:               name,
		config:             c,
		formatterName:      conf.FormatterName,
		transporterManager: transporter_manager.GetTransporterManager(""),
	}

	return a, nil
}
