package filelog

import (
	"fmt"
	"github.com/eolinker/eosc"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	"reflect"
)

const (
	driverName = "filelog"
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

//ConfigType 返回filelog驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建filelog驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, ok := v.(*DriverConfig)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(v))
	}
	c, err := ToConfig(conf)
	if err != nil {
		return nil, err
	}
	a := &filelog{
		id:                 id,
		name:               name,
		config:             c,
		formatterName:      conf.FormatterName,
		transporterManager: transporter_manager.GetLogTransporterManager(),
	}

	return a, nil
}

//accessDriver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type accessDriver struct {
	profession string
	name       string
	driver     string
	label      string
	desc       string
	configType reflect.Type
	params     map[string]string
}

//ConfigType 返回access-filelog驱动配置的反射类型
func (a *accessDriver) ConfigType() reflect.Type {
	return a.configType
}

//Create 创建access-filelog驱动实例
func (a *accessDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, ok := v.(*DriverConfigAccess)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfigAccess)(nil)), eosc.TypeNameOf(v))
	}

	c, err := ToConfigAccess(conf)
	if err != nil {
		return nil, err
	}

	worker := &filelogAccess{
		id:                 id,
		name:               name,
		config:             c,
		formatterName:      conf.FormatterName,
		fields:             conf.Fields,
		transporterManager: transporter_manager.GetAccessLogTransporterManager(),
	}

	return worker, nil
}
