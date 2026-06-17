package ai_proxy

import (
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
	"regexp"
)

type Config struct {
	ModelType       string            `json:"model_type"`
	Labels          map[string]string `json:"labels"`
	ModelIdFrom     string            `json:"model_id_from"`    // "path" | "body"
	ModelIdKey      string            `json:"model_id_key"`     // json key or path regex
	DefaultProvider string            `json:"default_provider"` // 默认供应商（当提取到的模型不含/时用于兜底）
}

func check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.ModelType == "" {
		cfg.ModelType = string(ai_convert.ModelTypeOpenAIChat)
	}

	if !ai_convert.ModelTypeIsVaild(ai_convert.ModelType(cfg.ModelType)) {
		return fmt.Errorf("model_type %s is not valid", cfg.ModelType)
	}

	if cfg.ModelIdFrom == "" {
		cfg.ModelIdFrom = "body"
	}

	if cfg.ModelIdFrom != "body" && cfg.ModelIdFrom != "path" {
		return fmt.Errorf("model_id_from %s is not valid, must be 'body' or 'path'", cfg.ModelIdFrom)
	}

	if cfg.ModelIdFrom == "body" && cfg.ModelIdKey == "" {
		cfg.ModelIdKey = "model"
	}

	if cfg.ModelIdFrom == "path" && cfg.ModelIdKey != "" {
		_, err := regexp.Compile(cfg.ModelIdKey)
		if err != nil {
			return fmt.Errorf("invalid path regex %s: %v", cfg.ModelIdKey, err)
		}
	}

	return nil
}
