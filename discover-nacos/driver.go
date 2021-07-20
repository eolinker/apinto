package discover_nacos

import (
	"errors"
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

func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, errors.New(fmt.Sprintf("error struct type: %s, need struct type: %s", eosc.TypeNameOf(v), d.configType))
	}
	return &nacos{
		id:             id,
		name:           name,
		address:        cfg.Config.Address,
		params:         cfg.Config.Params,
		labels:         cfg.Labels,
		services:       discovery.NewServices(),
		context:        nil,
		cancelFunc:     nil,
	}, nil

}


