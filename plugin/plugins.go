package plugin

import (
	"fmt"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/variable"
	"reflect"
	"sync"
)

var (
	pluginManger IPluginManager
	ones         sync.Once
)

type Plugins map[string]*Config

func (p Plugins) Reset(originVal reflect.Value, targetVal reflect.Value, variables map[string]string) ([]string, error) {
	ones.Do(func() {
		bean.Autowired(&pluginManger)
	})
	if originVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("plugin map reset error:%w %s", variable.ErrorUnsupportedKind, originVal.Kind())
	}
	targetType := targetVal.Type()
	newMap := reflect.MakeMap(targetType)
	usedVariables := make([]string, 0, len(variables))
	for _, key := range originVal.MapKeys() {
		// 判断是否存在对应的插件配
		cfgType, ok := pluginManger.GetConfigType(key.String())
		if !ok {
			return nil, fmt.Errorf("plugin %s not found", key.String())
		}
		value := originVal.MapIndex(key)
		newValue := reflect.New(targetType.Elem())

		used, err := pluginConfigSet(value, newValue, variables, cfgType)
		if err != nil {
			return nil, err
		}
		usedVariables = append(usedVariables, used...)
		newMap.SetMapIndex(key, newValue.Elem())
	}
	targetVal.Set(newMap)
	return usedVariables, nil
}

func pluginConfigSet(originVal reflect.Value, targetVal reflect.Value, variables map[string]string, cfgType reflect.Type) ([]string, error) {
	if targetVal.Kind() == reflect.Ptr {
		if !targetVal.Elem().IsValid() {
			targetType := targetVal.Type()
			newVal := reflect.New(targetType.Elem())
			targetVal.Set(newVal)
		}
		targetVal = targetVal.Elem()
	}
	usedVariables := make([]string, 0, len(variables))
	switch targetVal.Kind() {
	case reflect.Struct:
		{
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
				switch originVal.Elem().Kind() {
				case reflect.Map:
					{
						tag := field.Tag.Get("json")
						value = originVal.Elem().MapIndex(reflect.ValueOf(tag))
					}
				default:
					value = originVal.Elem()
				}
				variables, err := variable.RecurseReflect(value, fieldValue, variables)
				if err != nil {
					return nil, err
				}
				usedVariables = append(usedVariables, variables...)
				targetVal.Field(i).Set(fieldValue.Elem())
			}
		}
	case reflect.Ptr:
		return pluginConfigSet(originVal, targetVal, variables, cfgType)
	}
	return usedVariables, nil
}
