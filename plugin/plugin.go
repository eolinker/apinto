package plugin

import "github.com/eolinker/goku/filter"

//Config 普通插件配置，在router、service、upstream的插件格式
type Config struct {
	Disable bool        `json:"disable"`
	Config  interface{} `json:"config"`
}
type IPlugin interface {
	filter.IChain
	Destroy()
}
type IPluginManager interface {
	CreateRouter(id string, conf map[string]*Config) IPlugin
	CreateService(id string, conf map[string]*Config) IPlugin
	CreateUpstream(id string, conf map[string]*Config) IPlugin
}
