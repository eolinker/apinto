package reader_yaml

import (
	"errors"
	"fmt"

	"github.com/eolinker/eosc"
)

//Item 配置项
type Item map[string]interface{}

//Id 获取ID
func (i Item) Id() (string, bool) {
	v, has := i["id"]
	if !has {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

//Name 获取name
func (i Item) Name() (string, bool) {
	v, has := i["name"]
	if !has {
		return "", false
	}
	name, ok := v.(string)
	return name, ok
}

//Driver 获取driver
func (i Item) Driver() (string, bool) {
	v, has := i["driver"]
	if !has {
		return "", false
	}
	driver, ok := v.(string)
	return driver, ok
}

//newStoreValue 新建storeValue实例
func (i Item) newStoreValue(profession string, now string) (*eosc.StoreValue, error) {
	name := ""
	id, ok := i.Id()
	if !ok {
		name, ok = i.Name()
		if !ok {
			return nil, errors.New(fmt.Sprintf("id,name not found in %s", profession))
		}
		id = fmt.Sprintf("%s@%s", name, profession)
	}

	data, err := eosc.MarshalBytes(i)
	if err != nil {
		return nil, err
	}
	driver, ok := i.Driver()
	if !ok {
		return nil, errors.New(fmt.Sprintf("driver not found in %s", profession))
	}
	return &eosc.StoreValue{
		Id:         id,
		Profession: profession,
		Name:       name,
		Driver:     driver,
		CreateTime: now,
		UpdateTime: now,
		IData:      data,
	}, nil
}

//Config yaml文件配置项
type Config struct {
	Include   []string `json:":include" yaml:":include"`
	Router    []Item   `json:"router" yaml:"router"`
	Service   []Item   `json:"service" yaml:"service"`
	Upstream  []Item   `json:"upstream" yaml:"upstream"`
	Discovery []Item   `json:"discovery" yaml:"discovery"`
	Auth      []Item   `json:"auth" yaml:"auth"`
}
