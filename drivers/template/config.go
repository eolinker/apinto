package template

import "github.com/eolinker/apinto/plugin"

type Config struct {
	Plugins plugin.Plugins `json:"plugins" label:"插件配置"`
}
