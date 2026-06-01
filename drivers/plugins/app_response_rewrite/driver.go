package app_response_rewrite

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &Executor{
		WorkerBase: drivers.Worker(id, name),
		response:   response.Parse(v.Response),
	}, nil
}
