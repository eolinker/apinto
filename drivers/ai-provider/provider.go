package ai_provider

import (
	"embed"
	"reflect"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type ModelType string

const (
	ModelTypeLLM           ModelType = "llm"
	ModelTypeTextEmbedding ModelType = "text-embedding"
	ModelTypeSpeech2Text   ModelType = "speech2text"
	ModelTypeModeration    ModelType = "moderation"
	ModelTypeTTS           ModelType = "tts"
)

const (
	ModeChat     Mode = "chat"
	ModeComplete Mode = "complete"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

type Provider struct {
	Provider            string   `json:"provider" yaml:"provider"`
	SupportedModelTypes []string `json:"supported_model_types" yaml:"supported_model_types"`
}

type Model struct {
	Model           string     `json:"model" yaml:"model"`
	ModelType       ModelType  `json:"model_type" yaml:"model_type"`
	ModelProperties *ModelMode `json:"model_properties" yaml:"model_properties"`
}

type ModelMode struct {
	Mode        string `json:"mode" yaml:"mode"`
	ContextSize int    `json:"context_size" yaml:"context_size"`
}

func LoadModels(providerContent []byte, dirFs embed.FS) (map[string]*Model, error) {
	var provider Provider
	err := yaml.Unmarshal(providerContent, &provider)
	if err != nil {
		return nil, err
	}
	models := make(map[string]*Model)
	for _, modelType := range provider.SupportedModelTypes {
		dirFiles, err := dirFs.ReadDir(modelType)
		if err != nil {
			// 未找到模型目录
			continue
		}
		for _, dirFile := range dirFiles {
			if dirFile.IsDir() || !strings.HasSuffix(dirFile.Name(), ".yaml") {
				continue
			}
			modelContent, err := dirFs.ReadFile(modelType + "/" + dirFile.Name())
			if err != nil {
				return nil, err
			}
			var m Model
			err = yaml.Unmarshal(modelContent, &m)
			if err != nil {
				return nil, err
			}
			models[m.Model] = &m
		}

	}
	return models, nil
}

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
			jsonTag := field.Tag.Get("json")
			if jsonTag == k {
				// 获取字段的值
				fieldVal := val.Field(i)

				// 如果字段不可设置，跳过
				if !fieldVal.CanSet() {
					continue
				}

				// 根据字段的类型，进行类型转换
				switch fieldVal.Kind() {
				case reflect.Float64:
					if strVal, ok := v.(string); ok && strVal != "" {
						// 如果是 string 类型且非空，转换为 float64
						if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
							fieldVal.SetFloat(floatVal)
						}
					} else if floatVal, ok := v.(float64); ok {
						fieldVal.SetFloat(floatVal)
					}

				case reflect.Int:
					if intVal, ok := v.(int); ok {
						fieldVal.SetInt(int64(intVal))
					} else if strVal, ok := v.(string); ok && strVal != "" {
						if intVal, err := strconv.Atoi(strVal); err == nil {
							fieldVal.SetInt(int64(intVal))
						}
					} else if floatVal, ok := v.(float64); ok {
						fieldVal.SetInt(int64(floatVal))
					}
				case reflect.String:
					if strVal, ok := v.(string); ok {
						fieldVal.SetString(strVal)
					}

				default:
					// 其他类型不进行转换
				}
			}
		}
	}

	return &result
}
