package failover_strategy

import (
	"encoding/json"
	"fmt"

	"github.com/eolinker/apinto/strategy"
)

const (
	TriggerTypeDirect    = "direct"
	TriggerTypeCondition = "condition"
)

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop" label:"禁用"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Template    string                `json:"template" label:"模版ID"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	TriggerType string                `json:"trigger_type" label:"触发类型" enum:"direct,condition"`
	Triggers    TriggersConf          `json:"triggers" label:"触发器"`
	Providers   []*ProviderConf       `json:"providers" label:"灾备供应商列表"`
}

type TriggersConf struct {
	Failure TriggerFailureConf `json:"failure" label:"失败触发器"`
	Timeout TriggerTimeoutConf `json:"timeout" label:"超时触发器"`
}

type TriggerFailureConf struct {
	Enabled bool `json:"enabled" label:"是否启用"`
}

type TriggerTimeoutConf struct {
	Enabled        bool  `json:"enabled" label:"是否启用"`
	TimeoutSeconds int64 `json:"timeout_seconds" label:"超时时间(秒)"`
}

type ProviderConf struct {
	Name   string      `json:"name" label:"供应商名称"`
	Config *BaseConfig `json:"config" label:"供应商配置"`
}

type BaseConfig struct {
	BaseUrl string `json:"base_url"`
	APIKey  string `json:"apikey"`
}

func (b *BaseConfig) UnmarshalJSON(data []byte) error {
	var aux struct {
		BaseUrl1 string `json:"base_url"`
		BaseUrl2 string `json:"baseUrl"`
		APIKey1  string `json:"apikey"`
		APIKey2  string `json:"api_key"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.BaseUrl1 != "" {
		b.BaseUrl = aux.BaseUrl1
	} else {
		b.BaseUrl = aux.BaseUrl2
	}
	if aux.APIKey1 != "" {
		b.APIKey = aux.APIKey1
	} else {
		b.APIKey = aux.APIKey2
	}
	return nil
}

func (b BaseConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"base_url": b.BaseUrl,
		"baseUrl":  b.BaseUrl,
		"apikey":   b.APIKey,
		"api_key":  b.APIKey,
	})
}

func checkConfig(conf *Config) error {
	if conf.Priority < 1 {
		conf.Priority = 1
	} else if conf.Priority > 999 {
		return fmt.Errorf("priority value %d not allowed, must be between 1 and 999", conf.Priority)
	}

	if conf.TriggerType == "" {
		if conf.Triggers.Failure.Enabled || conf.Triggers.Timeout.Enabled {
			conf.TriggerType = TriggerTypeCondition
		} else {
			conf.TriggerType = TriggerTypeDirect
		}
	} else if conf.TriggerType != TriggerTypeDirect && conf.TriggerType != TriggerTypeCondition {
		return fmt.Errorf("unsupported trigger_type: %s, must be direct or condition", conf.TriggerType)
	}

	if conf.TriggerType == TriggerTypeCondition {
		if !conf.Triggers.Failure.Enabled && !conf.Triggers.Timeout.Enabled {
			return fmt.Errorf("at least one trigger (failure or timeout) must be enabled when trigger_type is condition")
		}
		if conf.Triggers.Timeout.Enabled && conf.Triggers.Timeout.TimeoutSeconds <= 0 {
			conf.Triggers.Timeout.TimeoutSeconds = 60
		}
	}

	if len(conf.Providers) == 0 {
		return fmt.Errorf("providers cannot be empty")
	}

	for i, p := range conf.Providers {
		if p.Name == "" {
			return fmt.Errorf("provider name cannot be empty at index %d", i)
		}
	}

	return nil
}
