package auth

import (
	"fmt"
	"reflect"

	"github.com/eolinker/apinto/auth"
	"github.com/eolinker/eosc"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	workers    eosc.IWorkers
	configType reflect.Type
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	_, err := d.check(v)
	if err != nil {
		return err
	}
	return nil
}
func (d *Driver) check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	return conf, nil
}
func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) getList(auths []eosc.RequireId) ([]auth.IAuth, error) {
	ls := make([]auth.IAuth, 0, len(auths))
	for _, id := range auths {
		worker, has := d.workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}
		if !worker.CheckSkill(auth.AuthSkill) {
			return nil, fmt.Errorf("%s:%w:%s", id, eosc.ErrorTargetNotImplementSkill, auth.AuthSkill)
		}
		ls = append(ls, worker.(auth.IAuth))

	}
	return ls, nil
}
func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}
	list, err := d.getList(conf.Auth)
	if err != nil {
		return nil, err
	}
	au := &Auth{
		Driver: d,
		id:     id,
		auths:  list,
	}

	return au, nil
}
