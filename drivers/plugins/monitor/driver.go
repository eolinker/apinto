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

func getList(ids []eosc.RequireId) ([]monitor_entry.IOutput, error) {
	ls := make([]monitor_entry.IOutput, 0, len(ids))
	for _, id := range ids {
		w, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		v, ok := w.(monitor_entry.IOutput)
		if !ok {
			return nil, fmt.Errorf("%s:worker d not implement IEntryOutput,now %v", string(id), reflect.TypeOf(w))
		}

		ls = append(ls, v)

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
		proxy := scope_manager.NewProxy(list...)

		monitorManager.SetProxyOutput(id, proxy)
	} else {
		monitorManager.SetProxyOutput(id, scope_manager.Get[monitor_entry.IOutput]("monitor"))
	}

	return o, nil
}
