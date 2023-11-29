package body_check

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Config struct {
	IsEmpty            bool `json:"is_empty" label:"是否允许为空"`
	AllowedPayloadSize int  `json:"allowed_payload_size" label:"允许的最大请求体大小"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	bc := &BodyCheck{
		WorkerBase:         drivers.Worker(id, name),
		isEmpty:            conf.IsEmpty,
		allowedPayloadSize: conf.AllowedPayloadSize * 1024,
	}

	return bc, nil
}
