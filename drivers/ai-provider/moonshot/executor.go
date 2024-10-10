package moonshot

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc/log"
	"reflect"
	"strconv"

	"github.com/eolinker/apinto/drivers"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed moonshot.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := ai_provider.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if v, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = v
			}
		}
	}
}

type Converter struct {
	apikey         string
	balanceHandler eocontext.BalanceHandler
	converter      convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apikey)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}

type executor struct {
	drivers.WorkerBase
	apikey string
	eocontext.BalanceHandler
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter, apikey: e.apikey}, true
}

func (e *executor) GetModel(model string) (convert.FGenerateConfig, bool) {
	if _, ok := modelConvert[model]; !ok {
		return nil, false
	}
	return func(cfg string) (map[string]interface{}, error) {

		result := map[string]interface{}{
			"model": model,
		}
		if cfg != "" {
			tmp := make(map[string]interface{})
			if err := json.Unmarshal([]byte(cfg), &tmp); err != nil {
				log.Errorf("unmarshal config error: %v, cfg: %s", err, cfg)
				return result, nil
			}
			modelCfg := mapToStruct[ModelConfig](tmp)
			result["frequency_penalty"] = modelCfg.FrequencyPenalty
			if modelCfg.MaxTokens >= 1 {
				result["max_tokens"] = modelCfg.MaxTokens
			}

			result["presence_penalty"] = modelCfg.PresencePenalty
			result["temperature"] = modelCfg.Temperature
			result["top_p"] = modelCfg.TopP
			if modelCfg.ResponseFormat == "" {
				modelCfg.ResponseFormat = "text"
			}
			result["response_format"] = map[string]interface{}{
				"type": modelCfg.ResponseFormat,
			}
		}
		return result, nil
	}, true
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config")
	}

	return e.reset(cfg, workers)
}

func (e *executor) reset(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	e.BalanceHandler = nil
	e.apikey = conf.APIKey
	convert.Set(e.Id(), e)

	return nil
}

func (e *executor) Stop() error {
	e.BalanceHandler = nil
	convert.Del(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}

type ModelConfig struct {
	FrequencyPenalty float64 `json:"frequency_penalty"`
	MaxTokens        int     `json:"max_tokens"`
	PresencePenalty  float64 `json:"presence_penalty"`
	ResponseFormat   string  `json:"response_format"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
}

func mapToStruct[T any](tmp map[string]interface{}) *T {
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
