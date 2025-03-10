package ai_convert

import (
	"reflect"
	"strconv"

	eoscContext "github.com/eolinker/eosc/eocontext"
)

type IProvider interface {
	Provider() string
	Model() string
	ModelConfig() map[string]interface{}
	Priority() int
	Health() bool
	Down()
	BalanceHandler() eoscContext.BalanceHandler
	GenExtender(cfg string) (map[string]interface{}, error)
}

//type ModelType string
//
//const (
//	ModelTypeLLM           ModelType = "llm"
//	ModelTypeTextEmbedding ModelType = "text-embedding"
//	ModelTypeSpeech2Text   ModelType = "speech2text"
//	ModelTypeModeration    ModelType = "moderation"
//	ModelTypeTTS           ModelType = "tts"
//)
//
//const (
//	ModeChat       Mode = "chat"
//	ModeCompletion Mode = "completion"
//)
//
//type Mode string
//
//func (m Mode) String() string {
//	return string(m)
//}
//
//type Provider struct {
//	Provider            string   `json:"provider" yaml:"provider"`
//	SupportedModelTypes []string `json:"supported_model_types" yaml:"supported_model_types"`
//}
//
//type Model struct {
//	Model           string     `json:"model" yaml:"model"`
//	ModelType       ModelType  `json:"model_type" yaml:"model_type"`
//	ModelProperties *ModelMode `json:"model_properties" yaml:"model_properties"`
//}
//
//type ModelMode struct {
//	Mode        string `json:"mode" yaml:"mode"`
//	ContextSize int    `json:"context_size" yaml:"context_size"`
//}
//
//func LoadModels(providerContent []byte, dirFs embed.FS) (map[string]*Model, error) {
//	var provider Provider
//	err := yaml.Unmarshal(providerContent, &provider)
//	if err != nil {
//		return nil, err
//	}
//	models := make(map[string]*Model)
//	for _, modelType := range provider.SupportedModelTypes {
//		dirFiles, err := dirFs.ReadDir(modelType)
//		if err != nil {
//			// 未找到模型目录
//			continue
//		}
//		for _, dirFile := range dirFiles {
//			if dirFile.IsDir() || !strings.HasSuffix(dirFile.Name(), ".yaml") {
//				continue
//			}
//			modelContent, err := dirFs.ReadFile(modelType + "/" + dirFile.Name())
//			if err != nil {
//				return nil, err
//			}
//			var m Model
//			err = yaml.Unmarshal(modelContent, &m)
//			if err != nil {
//				return nil, err
//			}
//			models[m.Model] = &m
//		}
//
//	}
//	return models, nil
//}

// MapToStruct 将 map 转换为结构体实例
func MapToStruct[T any](tmp map[string]interface{}) *T {
	// 创建目标结构体的实例
	var result T
	val := reflect.ValueOf(&result).Elem()

	// 获取结构体的类型
	t := val.Type()

	// 遍历 map 中的键值对
	for k, v := range tmp {
		// 查找结构体中与键名匹配的字段
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// 匹配字段的 json 标签或字段名
			jsonTag := field.Tag.Get("json")
			if jsonTag == k || field.Name == k {
				// 获取字段的值
				fieldVal := val.Field(i)

				// 如果字段不可设置，跳过
				if !fieldVal.CanSet() {
					continue
				}

				// 根据字段的类型，进行类型转换
				setValue(fieldVal, v)
			}
		}
	}

	return &result
}

// setValue 根据字段类型设置字段的值
func setValue(fieldVal reflect.Value, v interface{}) {
	switch fieldVal.Kind() {
	case reflect.Float64, reflect.Float32:
		setFloat(fieldVal, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		setInt(fieldVal, v)
	case reflect.String:
		if strVal, ok := v.(string); ok {
			fieldVal.SetString(strVal)
		}
	case reflect.Bool:
		setBool(fieldVal, v)
	case reflect.Slice:
		setSlice(fieldVal, v)
	case reflect.Struct:
		setStruct(fieldVal, v)
	default:
		// 其他类型不处理
	}
}

// setFloat 设置浮点数类型的字段值
func setFloat(fieldVal reflect.Value, v interface{}) {
	switch val := v.(type) {
	case float64:
		fieldVal.SetFloat(val)
	case float32:
		fieldVal.SetFloat(float64(val))
	case string:
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			fieldVal.SetFloat(floatVal)
		}
	}
}

// setInt 设置整数类型的字段值
func setInt(fieldVal reflect.Value, v interface{}) {
	switch val := v.(type) {
	case int:
		fieldVal.SetInt(int64(val))
	case int8, int16, int32, int64:
		fieldVal.SetInt(reflect.ValueOf(val).Int())
	case float64:
		fieldVal.SetInt(int64(val))
	case string:
		if intVal, err := strconv.Atoi(val); err == nil {
			fieldVal.SetInt(int64(intVal))
		}
	}
}

// setBool 设置布尔类型的字段值
func setBool(fieldVal reflect.Value, v interface{}) {
	switch val := v.(type) {
	case bool:
		fieldVal.SetBool(val)
	case string:
		if val == "true" {
			fieldVal.SetBool(true)
		} else if val == "false" {
			fieldVal.SetBool(false)
		}
	}
}

// setSlice 设置切片类型的字段值
func setSlice(fieldVal reflect.Value, v interface{}) {
	if reflect.TypeOf(v).Kind() == reflect.Slice {
		sliceVal := reflect.ValueOf(v)

		// 创建与目标字段类型匹配的切片
		sliceType := fieldVal.Type().Elem()
		newSlice := reflect.MakeSlice(reflect.SliceOf(sliceType), sliceVal.Len(), sliceVal.Len())

		// 遍历并设置切片元素
		for i := 0; i < sliceVal.Len(); i++ {
			elemVal := sliceVal.Index(i)
			newElem := reflect.New(sliceType).Elem()
			setValue(newElem, elemVal.Interface())
			newSlice.Index(i).Set(newElem)
		}

		fieldVal.Set(newSlice)
	}
}

// setStruct 设置结构体类型的字段值
func setStruct(fieldVal reflect.Value, v interface{}) {
	if nestedMap, ok := v.(map[string]interface{}); ok {
		// 创建嵌套结构体并递归映射
		nestedStruct := reflect.New(fieldVal.Type()).Elem()
		for i := 0; i < fieldVal.NumField(); i++ {
			field := fieldVal.Type().Field(i)
			jsonTag := field.Tag.Get("json")
			if val, exists := nestedMap[jsonTag]; exists {
				setValue(nestedStruct.Field(i), val)
			}
		}
		fieldVal.Set(nestedStruct)
	}
}
