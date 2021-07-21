package servicehttp

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

//ConfigType 返回service_http驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建service_http驱动的实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(v), eosc.ErrorStructType)
	}
	if work, has := workers[cfg.Upstream]; has {
		w := &serviceWorker{
			id:         id,
			name:       name,
			driver:     cfg.Driver,
			desc:       cfg.Desc,
			timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
			rewriteUrl: cfg.RewriteURL,
			retry:      cfg.Retry,
			scheme:     cfg.Scheme,
			upstream:   work.(upstream.IUpstream),
		}
		return w, nil
	}

	if work, has := workers[eosc.RequireId(fmt.Sprintf("%s@%s", cfg.Upstream, "upstream"))]; has {
		w := &serviceWorker{
			id:         id,
			name:       name,
			driver:     cfg.Driver,
			desc:       cfg.Desc,
			timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
			rewriteUrl: cfg.RewriteURL,
			retry:      cfg.Retry,
			scheme:     cfg.Scheme,
			upstream:   work.(upstream.IUpstream),
		}
		return w, nil
	}

	return nil, errors.New("fail to create serviceWorker")
}
