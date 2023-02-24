package prometheus

import (
	"fmt"
	metric_entry "github.com/eolinker/apinto/output"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return doCheck(v)
}

func check(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	err := doCheck(cfg)
	return cfg, err
}

func doCheck(cfg *Config) error {
	if len(cfg.Metrics) == 0 {
		return errNullMetric
	}
	return nil
}

func getList(ids []eosc.RequireId) ([]interface{}, error) {
	ls := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		worker, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		_, ok := worker.(metric_entry.IMetrics)
		if !ok {
			return nil, fmt.Errorf(errNotImpEntryFormat, string(id))
		}

		ls = append(ls, worker)

	}
	return ls, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	list, err := getList(conf.Output)
	if err != nil {
		return nil, err
	}

	p := &prometheus{
		WorkerBase: drivers.Worker(id, name),
		metrics:    conf.Metrics,
	}
	if len(list) > 0 {
		proxy := scope_manager.NewProxy()
		proxy.Set(list)
		p.proxy = proxy
	} else {
		p.proxy = scopeManager.Get(globalScopeName)
	}

	return p, nil
}
