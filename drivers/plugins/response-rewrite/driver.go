package response_rewrite

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return conf.doCheck()
}

func check(v interface{}) (*Config, error) {

	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	//若body非空且需要base64转码
	if conf.Body != "" && conf.BodyBase64 {
		conf.Body, err = utils.B64DecodeString(conf.Body)
		if err != nil {
			return nil, err
		}
	}

	r := &ResponseRewrite{
		WorkerBase: drivers.Worker(id, name),
		statusCode: conf.StatusCode,
		body:       conf.Body,
		headers:    conf.Headers,
		match:      conf.Match,
	}

	return r, nil
}
