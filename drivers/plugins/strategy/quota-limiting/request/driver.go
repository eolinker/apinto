package request

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Cache eosc.RequireId `json:"cache" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" label:"缓存位置"`
	Key   string         `json:"key"`
}

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	
	return &Strategy{
		
		WorkerBase: drivers.Worker(id, name),
		redisID:    string(cfg.Cache),
	}, nil
}
func CheckConfig(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.Key == "" {
		cfg.Key = "{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"
	}
	return nil
}
