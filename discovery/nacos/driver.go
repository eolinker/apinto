package nacos

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
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
		return nil, fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(v))
	}
	return &nacos{
		id:       id,
		name:     name,
		client:   newClient(cfg.Config.Address, cfg.getScheme()),
		nodes:    discovery.NewNodesData(),
		services: discovery.NewServices(),
		locker:   sync.RWMutex{},
	}, nil

}
