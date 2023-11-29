package proxy_rewrite_v2

import (
	"fmt"
	"regexp"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return v.doCheck()
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
		WorkerBase:  drivers.Worker(id, name),
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
