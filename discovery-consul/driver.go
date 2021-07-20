package discovery_consul

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
	"reflect"
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

func NewDriver() *driver {
	return &driver{configType: reflect.TypeOf(new(Config))}
}

func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	workerConfig, ok := v.(*Config)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(v), eosc.ErrorStructType)
	}
	return &consul{
		id:         id,
		name:       name,
		address:    workerConfig.Config.Address,
		params:     workerConfig.Config.Params,
		labels:     workerConfig.Labels,
		services:   discovery.NewServices(),
		context:    nil,
		cancelFunc: nil,
	}, nil
}
