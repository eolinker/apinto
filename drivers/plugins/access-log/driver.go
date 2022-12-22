package access_log

import (
	"fmt"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"

	"github.com/eolinker/eosc"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return nil
}

func check(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return cfg, nil
}

func getList(ids []eosc.RequireId) ([]interface{}, error) {
	ls := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		worker, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		_, ok := worker.(output.IEntryOutput)
		if !ok {
			return nil, fmt.Errorf("%s:worker not implement IEntryOutput", string(id))
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

	o := &accessLog{
		WorkerBase: drivers.Worker(id, name),
	}
	if len(list) > 0 {
		proxy := scope_manager.NewProxy()
		proxy.Set(list)
		o.proxy = proxy
	} else {
		o.proxy = scopeManager.Get("access_log")
	}

	return o, nil
}
