package ai_convert

import (
	"fmt"
	
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type IConverterCreateFunc func(cfg string) (IConverter, error)

type IConvertDriverCreateFunc[T any] func(ModelType, *T) (IConverterDriver, error)

type ModelType string

func (m ModelType) String() string {
	return string(m)
}

const (
	//ModelTypeChat 原生Chat格式
	ModelTypeChat ModelType = "chat"
	//ModelTypeOpenAIChat openAI兼容格式
	ModelTypeOpenAIChat ModelType = "openai-chat"
	//ModelTypeImageGeneration 图片生成（文生图）
	ModelTypeImageGeneration ModelType = "image-generation"
	//ModelTypeImageEdit 图片编辑（图生图）
	ModelTypeImageEdit ModelType = "image-edit"
	//ModelTypeImageTaskCommit 提交图片生成任务
	ModelTypeImageTaskCommit = "image-task-commit"
	//ModelTypeImageTaskQuery 查询图片
	ModelTypeImageTaskQuery = "image-task-query"
	//ModelTypeVideoTaskCommit 提交视频生成任务
	ModelTypeVideoTaskCommit ModelType = "video-task-commit"
	//ModelTypeVideoTaskQuery 查询视频
	ModelTypeVideoTaskQuery ModelType = "video-task-query"
)

var validModelType = map[ModelType]struct{}{
	ModelTypeChat:            {},
	ModelTypeOpenAIChat:      {},
	ModelTypeImageGeneration: {},
	ModelTypeVideoTaskCommit: {},
	ModelTypeVideoTaskQuery:  {},
	ModelTypeImageEdit:       {},
	ModelTypeImageTaskCommit: {},
	ModelTypeImageTaskQuery:  {},
}

func ModelTypeIsVaild(modelType ModelType) bool {
	_, ok := validModelType[modelType]
	return ok
}

type IConverterFactory interface {
	Create(cfg string) (IConverter, error)
}

type IConverter interface {
	Get(modelType ModelType) (IConverterDriver, bool)
	ModelTypeList() []ModelType
}

type IConverterDriver interface {
	Provider() string
	ModelType() ModelType
	RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error
	ResponseConvert(ctx eocontext.EoContext) error
}

type IChildConverter interface {
	IConverterDriver
	Endpoint() string
}
type FGenerateConfig func(cfg string) (map[string]interface{}, error)

func CheckKeySourceSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.key.IKeyResource"
}

func CheckProviderSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.provider.IProvider"
}

func NewConverter[T any](provider string, cfg *T, fns map[ModelType]IConvertDriverCreateFunc[T]) (IConverter, error) {
	if len(fns) == 0 {
		return nil, fmt.Errorf("no driver found for %s provider", provider)
	}
	c := &Converter{
		drivers: eosc.BuildUntyped[ModelType, IConverterDriver](),
	}
	for mt, fn := range fns {
		driver, err := fn(mt, cfg)
		if err != nil {
			return nil, err
		}
		c.drivers.Set(mt, driver)
	}
	
	return c, nil
}

type Converter struct {
	drivers eosc.Untyped[ModelType, IConverterDriver]
}

func (c *Converter) Get(modelType ModelType) (IConverterDriver, bool) {
	if c.drivers == nil {
		return nil, false
	}
	return c.drivers.Get(modelType)
}

func (c *Converter) ModelTypeList() []ModelType {
	return c.drivers.Keys()
}
