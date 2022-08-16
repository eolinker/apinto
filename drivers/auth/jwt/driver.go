package jwt

import (
	"reflect"

	"github.com/eolinker/eosc"
)

const (
	driverName = "jwt"
)

//driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type driver struct {
	profession string
	name       string
	driver     string
	label      string
	desc       string
	configType reflect.Type
}

//ConfigType 返回jwt鉴权驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建jwt鉴权驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	a := &jwt{
		id: id,
	}
	err := a.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return a, nil
}
