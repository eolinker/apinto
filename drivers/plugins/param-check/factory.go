package param_check

import (
	"fmt"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "param_check"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	headerChecker := make([]*paramChecker, 0)
	queryChecker := make([]*paramChecker, 0)
	bodyChecker := make([]*paramChecker, 0)
	for _, v := range conf.Params {
		c, err := checker.Parse(v.MatchText)
		if err != nil {
			return nil, fmt.Errorf("parse param check text error: %w,text: %s", err, v.MatchText)
		}
		switch v.Position {
		case "header":
			headerChecker = append(headerChecker, &paramChecker{
				name:    v.Name,
				Checker: c,
			})
		case "query":
			queryChecker = append(queryChecker, &paramChecker{
				name:    v.Name,
				Checker: c,
			})
		case "body":
			bodyChecker = append(bodyChecker, &paramChecker{
				name:    v.Name,
				Checker: c,
			})
		}
	}

	return &executor{
		WorkerBase:    drivers.Worker(id, name),
		headerChecker: headerChecker,
		queryChecker:  queryChecker,
		bodyChecker:   bodyChecker,
	}, nil
}
