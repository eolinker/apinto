package response_filter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name            = "response_filter"
	WhiteFilterType = "white"
	BlackFilterType = "black"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, check)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	filters := make([]IFilter, 0, 4)
	if len(conf.BodyFilter) > 0 {
		switch conf.BodyFilterType {
		case WhiteFilterType:
			filter, err := NewBodyWhiteFilter(conf.BodyFilter)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filter)
		case BlackFilterType:
			filter, err := NewBodyBlackFilter(conf.BodyFilter)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filter)
		}
	}

	if len(conf.HeaderFilter) > 0 {
		switch conf.HeaderFilterType {
		case WhiteFilterType:
			filter, err := NewHeaderWhiteFilter(conf.HeaderFilter)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filter)
		case BlackFilterType:
			filter, err := NewHeaderBlackFilter(conf.HeaderFilter)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filter)
		}
	}

	return &executor{
		WorkerBase: drivers.Worker(id, name),
		filters:    filters,
	}, nil
}
