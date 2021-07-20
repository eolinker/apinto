package nacos

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
	"reflect"
)

const (
	driverName = "nacos"
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

//ConfigType 返回nacos驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建nacos驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(v), eosc.ErrorStructType)
	}
	return &nacos{
		id:         id,
		name:       name,
		address:    cfg.Config.Address,
		params:     cfg.Config.Params,
		labels:     cfg.Labels,
		services:   discovery.NewServices(),
		context:    nil,
		cancelFunc: nil,
	}, nil

}
