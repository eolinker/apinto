package monitor

import (
	"fmt"
	"reflect"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	monitor_entry "github.com/eolinker/apinto/entries/monitor-entry"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
)

func getList(ids []eosc.RequireId) ([]interface{}, error) {
	ls := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		worker, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		_, ok := worker.(monitor_entry.IOutput)
		if !ok {
			return nil, fmt.Errorf("%s:worker d not implement IEntryOutput,now %v", string(id), reflect.TypeOf(worker))
		}

		ls = append(ls, worker)

	}
	return ls, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	log.Info("create monitor worker...")
	list, err := getList(conf.Output)
	if err != nil {
		return nil, err
	}

	o := &worker{
		WorkerBase: drivers.Worker(id, name),
	}
	if len(list) > 0 {
		proxy := scope_manager.NewProxy()
		proxy.Set(list)
		monitorManager.SetProxyOutput(id, proxy)
	} else {
		monitorManager.SetProxyOutput(id, scopeManager.Get("monitor"))
	}

	return o, nil
}
