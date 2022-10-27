package access_log

import (
	"fmt"
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

func getList(auths []eosc.RequireId) ([]output.IEntryOutput, error) {
	ls := make([]output.IEntryOutput, 0, len(auths))
	for _, id := range auths {
		worker, has := workers.Get(string(id))
		if !has {
			return nil, fmt.Errorf("%s:%w", id, eosc.ErrorWorkerNotExits)
		}

		outPut, ok := worker.(output.IEntryOutput)
		if !ok {
			return nil, fmt.Errorf("%s:worker not implement IEntryOutput", string(id))
		}

		ls = append(ls, outPut)

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
		output:     list,
	}

	return o, nil
}
