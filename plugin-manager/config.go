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
	Name       string                 `json:"name"`
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	Config     interface{}            `json:"config"`
	InitConfig map[string]interface{} `json:"init_config"`
}
