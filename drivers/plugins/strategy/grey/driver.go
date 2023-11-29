package grey

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Config struct {
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return &Strategy{
		WorkerBase: drivers.Worker(id, name),
	}, nil
}
