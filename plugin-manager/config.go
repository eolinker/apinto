package plugin_manager

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/variable"
	"reflect"
)

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"
)

type PluginWorkerConfig struct {
	Plugins []*PluginConfig `json:"plugins" yaml:"plugins"`
}

//PluginConfig 全局插件配置
type PluginConfig struct {
	Name       string                 `json:"name" yaml:"name" `
	ID         string                 `json:"id" yaml:"id"`
	Status     string                 `json:"status" yaml:"status"`
	Config     interface{}            `json:"config" yaml:"config"`
	InitConfig map[string]interface{} `json:"init_config" yaml:"init_config"`
}

func (p *PluginConfig) Reset(originVal reflect.Value, targetVal reflect.Value, params map[string]string, configTypes *eosc.ConfigType) ([]string, error) {
	if originVal.Kind() == reflect.Ptr {
		originVal = originVal.Elem()
	}
	if originVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("plugin map reset error:%w %s", variable.ErrorUnsupportedKind, originVal.Kind())
	}
	nameVal := originVal.MapIndex(reflect.ValueOf("name"))
	if !nameVal.IsValid() {
		// 当name字段不存在，则报错
		return nil, fmt.Errorf("missing field name")
	}
	cfgType, ok := configTypes.Get(nameVal.Elem().String())
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", nameVal.Elem().String())
	}
	usedVariables := make([]string, 0, len(params))
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
		used, err := variable.RecurseReflect(value, fieldValue, params, configTypes)
		if err != nil {
			return nil, err
		}
		usedVariables = append(usedVariables, used...)
		targetVal.Field(i).Set(fieldValue.Elem())
	}
	return usedVariables, nil
}
