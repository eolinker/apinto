package prometheus

import (
	"fmt"

	output "github.com/eolinker/apinto/output"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
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

func getList(ids []eosc.RequireId) ([]output.IMetrics, error) {
	ls := make([]output.IMetrics, 0, len(ids))
	for _, id := range ids {
		w, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		v, ok := w.(output.IMetrics)
		if !ok {
			return nil, fmt.Errorf(errNotImpEntryFormat, string(id))
		}

		ls = append(ls, v)

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
		proxy := scope_manager.NewProxy(list...)

		p.proxy = proxy
	} else {
		p.proxy = scope_manager.Get[output.IMetrics](globalScopeName)
	}

	return p, nil
}
