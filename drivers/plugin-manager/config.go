package plugin_manager

import (
	"fmt"
	"reflect"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/variable"
)

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"
)

type PluginWorkerConfig struct {
	Plugins []*PluginConfig `json:"plugins" yaml:"plugins"`
}

// PluginConfig 全局插件配置
type PluginConfig struct {
	Name       string                 `json:"name" yaml:"name" `
	ID         string                 `json:"id" yaml:"id"`
	Status     string                 `json:"status" yaml:"status"`
	Config     interface{}            `json:"config" yaml:"config"`
	InitConfig map[string]interface{} `json:"init_config" yaml:"init_config"`
}

func (p *PluginConfig) GetType(originVal reflect.Value) (reflect.Type, error) {
	idVal := originVal.MapIndex(reflect.ValueOf("id"))

	if !idVal.IsValid() {
		// 当id字段不存在，则报错
		return nil, fmt.Errorf("missing field name")
	}
	id := idVal.Elem().String()
	nameVal := originVal.MapIndex(reflect.ValueOf("id"))
	if !nameVal.IsValid() {
		// 当name字段不存在，则报错
		return nil, fmt.Errorf("missing field name")
	}
	name := nameVal.Elem().String()

	var params map[string]interface{} = nil
	paramsVal := originVal.MapIndex(reflect.ValueOf("init_config"))
	if paramsVal.IsValid() {
		tmp, ok := paramsVal.Elem().Interface().(map[string]interface{})
		if ok {
			params = tmp
		}
	}

	factory, has := singleton.extenderDrivers.GetDriver(id)
	if !has {
		return nil, fmt.Errorf("driver(%s) not found", id)
	}
	driver, err := factory.Create(id, name, name, name, params)
	if err != nil {
		return nil, fmt.Errorf("create driver(%s) error:%s", idVal.Elem().String(), err)
	}
	return driver.ConfigType(), nil
}

func (p *PluginConfig) Reset(originVal reflect.Value, targetVal reflect.Value, variables eosc.IVariable) ([]string, error) {
	if originVal.Kind() == reflect.Ptr {
		originVal = originVal.Elem()
	}
	if originVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("plugin map reset error:%w %s", eosc.ErrorUnsupportedKind, originVal.Kind())
	}

	cfgType, err := p.GetType(originVal)
	if err != nil {
		return nil, err
	}
	usedVariables := make([]string, 0, variables.Len())
	targetType := targetVal.Type()
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		var fieldValue reflect.Value
		switch field.Type.Kind() {
		case reflect.Interface:
			if cfgType.Kind() == reflect.Ptr {
				cfgType = cfgType.Elem()
			}
			fieldValue = reflect.New(cfgType)
		default:
			fieldValue = reflect.New(field.Type)
		}
		var value reflect.Value
		switch originVal.Kind() {
		case reflect.Map:
			{
				tag := field.Tag.Get("json")
				value = originVal.MapIndex(reflect.ValueOf(tag))
			}
		default:
			value = originVal
		}
		used, err := variable.RecurseReflect(value, fieldValue, variables)
		if err != nil {
			return nil, err
		}
		usedVariables = append(usedVariables, used...)
		targetVal.Field(i).Set(fieldValue.Elem())
	}
	return usedVariables, nil
}
