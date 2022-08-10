package proxy_rewrite_v2

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
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

func (d *Driver) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}

	pw := &ProxyRewrite{
		Driver:      d,
		id:          id,
		pathType:    conf.PathType,
		notMatchErr: conf.NotMatchErr,
		hostRewrite: conf.HostRewrite,
		host:        conf.Host,
		headers:     conf.Headers,
	}

	switch conf.PathType {
	case typeStatic:
		pw.staticPath = conf.StaticPath
	case typePrefix:
		pw.prefixPath = conf.PrefixPath
	case typeRegex:
		regexMatch := make([]*regexp.Regexp, 0)

		for _, rPath := range conf.RegexPath {
			rMatch, err := regexp.Compile(rPath.RegexPathMatch)
			if err != nil {
				return nil, fmt.Errorf(regexpErrInfo, rPath.RegexPathMatch)
			}
			regexMatch = append(regexMatch, rMatch)
		}
		pw.regexPath = conf.RegexPath
		pw.regexMatch = regexMatch
	}

	return pw, nil
}
