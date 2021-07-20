package yaml

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/store"
)

const (
	mode       = "yaml"
	keyInclude = ":include"
)

func Register() error {
	store.RegisterStoreFactory(mode, CreateStore)
	return nil
}

//CreateStore 创建store
func CreateStore(config map[string]string) (eosc.IStore, error) {
	store := NewStore()
	path, ok := config["conf"]
	if !ok {
		return nil, errors.New("conf path is empty")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	convertData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	var cfg Employees
	err = json.Unmarshal(convertData, &cfg)
	if err != nil {
		return nil, err
	}

	store.SetEmployee(cfg)
	return store, nil
}

type Employees map[string][]interface{}

//TODO
//func Register() {
//	stores.RegisterFactory(mode, newStoreYamlFactory())
//}
//
//type storeYamlFactory struct {
//}
//
//func newStoreYamlFactory() *storeYamlFactory {
//	return &storeYamlFactory{}
//}
//
////CreateStore 创建store
//func (s *storeYamlFactory) CreateStore(config map[string]string) (eosc.IStore, error) {
//	store := NewStore()
//	path, ok := config["conf"]
//	if !ok {
//		return nil, errors.New("conf path is empty")
//	}
//	data, err := ioutil.ReadFile(path)
//	if err != nil {
//		return nil, err
//	}
//
//	convertData, err := yaml.YAMLToJSON(data)
//	if err != nil {
//		return nil, err
//	}
//	var cfg Employees
//	err = json.Unmarshal(convertData, &cfg)
//	if err != nil {
//		return nil, err
//	}
//
//	store.SetEmployee(cfg)
//	return store, nil
//}
//
//type Employees map[string][]interface{}
