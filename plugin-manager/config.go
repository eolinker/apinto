package plugin_manager

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"

	pluginRouter   = "router"
	pluginService  = "service"
	pluginUpstream = "upstream"
)

type PluginWorkerConfig struct {
	Plugins []*PluginConfig
}

//PluginConfig 全局插件配置
type PluginConfig struct {
	Name   string      `json:"name"`
	ID     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Config interface{} `json:"config"`
}

//OrdinaryPlugin 普通插件配置，在router、service、upstream的插件格式
type OrdinaryPlugin struct {
	Disable bool        `json:"disable"`
	Config  interface{} `json:"config"`
}
