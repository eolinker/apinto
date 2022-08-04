package plugin

import (
	"errors"
	"github.com/eolinker/eosc/variable"
	"reflect"
)

var (
	ErrorVariableNotFound = errors.New("variable not found")
	ErrorUnsupportedKind  = errors.New("unsupported kind")
)

type Plugins map[string]*Config

func (p Plugins) Reset(originVal reflect.Value, targetVal reflect.Value, params map[string]string) error {
	//if originVal.Kind() != reflect.Map {
	//	return fmt.Errorf("plugin map reset error:%w %s", ErrorUnsupportedKind, originVal.Kind())
	//}
	//targetType := targetVal.Type()
	//newMap := reflect.MakeMap(targetType)
	//for _, key := range originVal.MapKeys() {
	//	// 判断是否存在对应的插件配置
	//	cfgType, ok := typeMap[key.String()]
	//	if !ok {
	//		return fmt.Errorf("plugin %s not found", key.String())
	//	}
	//	value := originVal.MapIndex(key)
	//	newValue := reflect.New(targetType.Elem())
	//
	//	err := pluginConfigSet(value, newValue, params, cfgType)
	//	if err != nil {
	//		return err
	//	}
	//	newMap.SetMapIndex(key, newValue.Elem())
	//}
	//targetVal.Set(newMap)
	return nil
}

func pluginConfigSet(originVal reflect.Value, targetVal reflect.Value, params map[string]string, cfgType reflect.Type) error {
	if targetVal.Kind() == reflect.Ptr {
		if !targetVal.Elem().IsValid() {
			targetType := targetVal.Type()
			newVal := reflect.New(targetType.Elem())
			targetVal.Set(newVal)
		}
		targetVal = targetVal.Elem()
	}
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

				err := variable.RecurseReflect(value, fieldValue, params)
				if err != nil {
					return err
				}
				targetVal.Field(i).Set(fieldValue.Elem())
			}
		}
	case reflect.Ptr:
		return pluginConfigSet(originVal, targetVal, params, cfgType)
	}
	return nil
}
