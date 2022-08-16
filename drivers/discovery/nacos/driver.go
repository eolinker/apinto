package nacos

import (
	"fmt"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
	"sync"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"
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
}

//ConfigType 返回nacos驱动配置的反射类型
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

//Create 创建nacos驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, ok := v.(*Config)
	if !ok {
		val := reflect.ValueOf(v)
		log.Debug("reflect", val.Kind(), val.Interface())
		return nil, fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(v))
	}
	return &nacos{
		id:       id,
		name:     name,
		client:   newClient(cfg.Config.Address, cfg.getParams()),
		nodes:    discovery.NewNodesData(),
		services: discovery.NewServices(),
		locker:   sync.RWMutex{},
	}, nil

}
