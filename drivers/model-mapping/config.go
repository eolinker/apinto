package model_mapping

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/strategy"
)

// Config 策略驱动配置结构
type Config struct {
	Name        string                `json:"name" label:"策略名称"`
	Description string                `json:"description" label:"策略描述"`
	Stop        bool                  `json:"stop" label:"是否禁用"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999，数值越小越优先"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Targets     []string              `json:"targets" label:"目标模型列表" description:"格式为 provider/model，第0位为主选，后续为灾备备选"`
}

// Target 解析后的目标模型对象
type Target struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Raw      string `json:"raw"`
}

// ParseTarget 解析目标字符串，格式要求为 {provider}/{model}
func ParseTarget(raw string) (*Target, error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("invalid target format '%s', expected '{provider}/{model}'", raw)
	}

	return &Target{
		Provider: strings.TrimSpace(parts[0]),
		Model:    strings.TrimSpace(parts[1]),
		Raw:      raw,
	}, nil
}

// Check 检查配置的合法性
func Check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type: %T", v)
	}

	//if conf.Priority < 1 || conf.Priority > 999 {
	//	return nil, fmt.Errorf("priority must be in [1, 999], got %d", conf.Priority)
	//}

	if len(conf.Targets) == 0 {
		return nil, fmt.Errorf("targets cannot be empty")
	}

	for _, t := range conf.Targets {
		if _, err := ParseTarget(t); err != nil {
			return nil, err
		}
	}

	if conf.Filters == nil {
		return nil, fmt.Errorf("filters cannot be nil")
	}

	// 必须包含 consumer 和 model 规则且非空
	if len(conf.Filters["consumer"]) == 0 {
		return nil, fmt.Errorf("filters.consumer cannot be empty")
	}
	if len(conf.Filters["model"]) == 0 {
		return nil, fmt.Errorf("filters.model cannot be empty")
	}

	return conf, nil
}
