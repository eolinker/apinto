package basic

import (
	"reflect"

	"github.com/eolinker/eosc"
)

const (
	driverName = "basic"
)

//driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type driver struct {
	profession string
	name       string

	label      string
	desc       string
	configType reflect.Type
}

//ConfigType 返回http_proxy驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建http_proxy驱动的实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	w := &basic{
		id: id,
	}
	err := w.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	return w, nil
}
