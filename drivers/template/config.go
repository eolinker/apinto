package template

import "github.com/eolinker/apinto/plugin"

type Config struct {
	Plugins map[string]*plugin.Config `json:"plugins" label:"插件配置"`
}
