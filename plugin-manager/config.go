package plugin_manager

import (
	"reflect"
)

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"
)

type PluginWorkerConfig struct {
	Plugins []*PluginConfig `json:"plugins" yaml:"plugins"`
}

//PluginConfig 全局插件配置
type PluginConfig struct {
	Name       string                 `json:"name" yaml:"name" `
	ID         string                 `json:"id" yaml:"id"`
	Status     string                 `json:"status" yaml:"status"`
	Config     interface{}            `json:"config" yaml:"config"`
	InitConfig map[string]interface{} `json:"init_config" yaml:"init_config"`
}

func (p *PluginConfig) Reset(originVal reflect.Value, targetVal reflect.Value, params map[string]string) error {
	return nil
}
