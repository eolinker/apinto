package limiting

import (
	"fmt"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/config"
	"reflect"
)

type Config struct {
	Cache eosc.RequireId `json:"cache" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" label:"缓存位置"`
}
type driver struct {
}

func (d *driver) ConfigType() reflect.Type {
	return configType
}

func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(v))
	}
	return &Strategy{
		id:    id,
		name:  name,
		cache: resources.NewCacheBuilder(string(cfg.Cache)),
	}, nil
}
