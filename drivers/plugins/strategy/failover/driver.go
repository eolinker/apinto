package failover

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	ModelType string `json:"model_type" label:"模型类型" description:"默认 openai/chat"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	modelType := conf.ModelType
	if modelType == "" {
		modelType = string(ai_convert.ModelTypeOpenAIChat)
	}
	return &Strategy{
		WorkerBase: drivers.Worker(id, name),
		modelType:  ai_convert.ModelType(modelType),
	}, nil
}
