package protocbuf

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// Create 创建service_http驱动的实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	parseFiles(v.ProtoFiles)

	return &Worker{
		WorkerBase: drivers.Worker(id, name),
		source:     nil,
	}, nil
}
