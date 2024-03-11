package js_inject

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Variables        []Variable `json:"variables" label:"变量列表"`
	InjectCode       string     `json:"inject" label:"注入代码"`
	MatchContentType []string   `json:"match_content_type" label:"匹配的Content-Type"`
}

type Variable struct {
	Key   string `json:"key" label:"变量名"`
	Value string `json:"value" label:"变量值"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	e := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	e.reset(conf)
	return e, nil
}
