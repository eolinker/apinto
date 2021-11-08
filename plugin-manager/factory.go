package plugin_manager

import "github.com/eolinker/eosc/http"

type IPluginFactory interface {
	Name() string
	Create(cfg interface{}) (http.IFilter, error)
}
