package cache

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var (
	workers eosc.IWorkers
)

func init() {
	bean.Autowired(&workers)
}

type Config struct {
	Cache eosc.RequireId `json:"cache" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" label:"缓存位置"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return &Strategy{
		WorkerBase: drivers.Worker(id, name),
		cache:      scope_manager.Auto[resources.ICache](string(conf.Cache), "redis"),
	}, nil
}
