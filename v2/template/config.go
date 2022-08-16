package template

import "github.com/eolinker/apinto/plugin"

type Config struct {
	plugins map[string]*plugin.Config `json:"plugins"`
}
