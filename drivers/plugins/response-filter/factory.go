package response_filter

import (
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/ohler55/ojg/jp"
)

const (
	Name = "response_filter"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	bodyFilter := make([]jp.Expr, 0, len(conf.BodyFilter))
	for _, filter := range conf.BodyFilter {
		key := filter
		if !strings.HasPrefix(key, "$.") {
			key = "$." + key
		}
		expr, err := jp.ParseString(filter)
		if err != nil {
			return nil, err
		}
		bodyFilter = append(bodyFilter, expr)
	}

	return &executor{
		WorkerBase:   drivers.Worker(id, name),
		bodyFilter:   bodyFilter,
		headerFilter: conf.HeaderFilter,
	}, nil
}
