package example

import (
	"errors"
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	configType reflect.Type
	params     map[string]string
}

func (h *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	config, ok := v.(*Config)
	if !ok {
		return nil, errors.New("error")
	}

	return NewExample(config, workers), nil
}
func (h *Driver) ConfigType() reflect.Type {

	return reflect.TypeOf(new(Config))

}
