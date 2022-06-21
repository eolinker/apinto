package proxy_rewrite

import (
	"fmt"
	"github.com/eolinker/eosc"
	"reflect"
	"regexp"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	_, err := d.check(v)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}

	err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}

	pw := &ProxyRewrite{
		Driver:   d,
		id:       id,
		scheme:   conf.Scheme,
		uri:      conf.URI,
		regexURI: conf.RegexURI,
		host:     conf.Host,
		headers:  conf.Headers,
	}

	if len(conf.RegexURI) > 0 {
		pw.regexMatch, err = regexp.Compile(conf.RegexURI[0])
		if err != nil {
			return nil, fmt.Errorf(regexpErrInfo, conf.RegexURI[0])
		}
	}

	return pw, nil
}
