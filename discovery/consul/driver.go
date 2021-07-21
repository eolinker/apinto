package consul

import (
	"reflect"

	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/eosc"
)

const (
	driverName = "consul"
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

//ConfigType 返回consul驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建consul驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	c := &consul{
		id:       id,
		name:     name,
		services: discovery.NewServices(),
	}
	err := c.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return c, nil
}
