package service_http

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
)

const (
	driverName = "http"
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
		return nil, errors.New(fmt.Sprintf("error struct type: %s, need struct type: %s", eosc.TypeNameOf(v), d.configType))
	}
	if work, has := workers[cfg.Upstream]; has {
		w := &serviceWorker{
			id:         id,
			name:       name,
			driver:     cfg.Driver,
			desc:       cfg.Desc,
			timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
			rewriteUrl: cfg.RewriteUrl,
			retry:      cfg.Retry,
			scheme:     cfg.Scheme,
			upstream:   work.(upstream.IUpstream),
		}
		return w, nil
	} else {
		work, has = workers[eosc.RequireId(fmt.Sprintf("%s@%s", cfg.Upstream, "upstream"))]
		if has {
			w := &serviceWorker{
				id:         id,
				name:       name,
				driver:     cfg.Driver,
				desc:       cfg.Desc,
				timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
				rewriteUrl: cfg.RewriteUrl,
				retry:      cfg.Retry,
				scheme:     cfg.Scheme,
				upstream:   work.(upstream.IUpstream),
			}
			return w, nil
		}
	}
	return nil, errors.New("fail to create serviceWorker")
}
