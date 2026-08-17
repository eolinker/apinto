package amount

import (
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Cache                eosc.RequireId `json:"cache" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" label:"缓存位置"`
	Key                  string         `json:"key"`
	PriceKey             string         `json:"price_key"`
	VersionKey           string         `json:"version_key"`
	BindResourceGroupKey string         `json:"bind_resource_group_key"`
}

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &Strategy{
		WorkerBase:           drivers.Worker(id, name),
		redisID:              string(cfg.Cache),
		key:                  context_label.NewKeyGenerator(cfg.Key),
		versionKey:           context_label.NewKeyGenerator(cfg.VersionKey),
		priceKey:             context_label.NewKeyGenerator(cfg.PriceKey),
		bindResourceGroupKey: context_label.NewKeyGenerator(cfg.BindResourceGroupKey),
	}, nil
}

func CheckConfig(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.Key == "" {
		cfg.Key = "{product}:quota-limiting:{strategy}:{target_type}:{application}:{dimension}:{time_format}"
	}
	if cfg.PriceKey == "" {
		cfg.PriceKey = "{product}:access-resource-price:{application}:{resource}"
	}
	if cfg.VersionKey == "" {
		cfg.VersionKey = "{product}:version:access-resource-price:{resource}:{version}"
	}
	if cfg.BindResourceGroupKey == "" {
		cfg.BindResourceGroupKey = "user_bind_resource_group:{application}"
	}
	return nil
}
