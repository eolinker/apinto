package grey

import (
	"github.com/eolinker/eosc"
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

	return &Strategy{
		id:   id,
		name: name,
	}, nil
}
