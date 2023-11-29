package proxy_rewrite

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	err := conf.doCheck()
	if err != nil {
		return err
	}
	return nil
}

func check(v interface{}) (*Config, error) {

	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	err = conf.doCheck()
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

	pw := &ProxyRewrite{
		WorkerBase: drivers.Worker(id, name),
		uri:        conf.URI,
		regexURI:   conf.RegexURI,
		host:       conf.Host,
		headers:    conf.Headers,
	}

	if len(conf.RegexURI) > 0 {
		pw.regexMatch, err = regexp.Compile(conf.RegexURI[0])
		if err != nil {
			return nil, fmt.Errorf(regexpErrInfo, conf.RegexURI[0])
		}
	}

	return pw, nil
}
