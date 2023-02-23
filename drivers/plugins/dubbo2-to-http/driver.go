package dubbo2_to_http

import (
	"errors"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	if conf.Method == "" {
		return nil, errors.New("method is null")
	}

	if len(conf.Params) == 0 {
		return nil, errors.New("params is null")
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	if conf.Method == "" {
		return nil, errors.New("method is null")
	}

	if len(conf.Params) == 0 {
		return nil, errors.New("params is null")
	}

	params := make([]param, 0, len(conf.Params))

	for _, p := range conf.Params {
		params = append(params, param{
			className: p.ClassName,
			fieldName: p.FieldName,
		})
	}

	pw := &ToHttp{
		WorkerBase:  drivers.Worker(id, name),
		method:      conf.Method,
		path:        conf.Path,
		contentType: conf.ContentType,
		params:      params,
	}

	return pw, nil
}
