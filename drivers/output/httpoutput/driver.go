package httpoutput

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return doCheck(v)
}
func doCheck(v *Config) error {

	httpConf := v
	if httpConf.Method == "" {
		return errMethod
	}
	switch httpConf.Method {
	case "GET", "POST", "HEAD", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE":
	default:
		return errMethod
	}

	if httpConf.Url == "" {
		return errUrlNull
	}

	if httpConf.Type == "" {
		httpConf.Type = "line"
	}

	switch httpConf.Type {
	case "line", "json":
	default:
		return errFormatterType
	}

	if len(httpConf.Formatter) == 0 {
		return errFormatterConf
	}

	return nil
}
func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	err = doCheck(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil

}
func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := doCheck(conf)
	if err != nil {
		return nil, err
	}
	worker := &HttpOutput{
		WorkerBase: drivers.Worker(id, name),
		config:     conf,
	}

	return worker, nil
}
