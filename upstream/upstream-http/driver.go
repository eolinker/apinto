package upstream_http

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/eosc"
)

const (
	driverName = "http_proxy"
)

var (
	ErrorStructType = "error struct type: %s, need struct type: %s"
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
		return nil, errors.New(fmt.Sprintf(ErrorStructType, eosc.TypeNameOf(v), d.configType))
	}
	if factory, has := workers[cfg.Discovery]; has {
		f, ok := factory.(discovery.IDiscovery)
		if ok {
			app, err := f.GetApp(cfg.Config)
			if err != nil {
				return nil, err
			}
			w := &httpUpstream{
				id:          id,
				name:        name,
				driver:      cfg.Driver,
				desc:        cfg.Desc,
				scheme:      cfg.Scheme,
				balanceType: cfg.Type,
				app:         app,
			}
			return w, nil
		}

	}
	return nil, errors.New("fail to create upstream worker")
}
