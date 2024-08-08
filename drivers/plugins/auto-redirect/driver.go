package auto_redirect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

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

	redirectCount := conf.MaxRedirectCount
	if redirectCount < 1 || redirectCount > maxRedirectCount {
		redirectCount = maxRedirectCount
	}
	redirectPrefix := ""
	if conf.PathPrefix != "" {
		u, err := url.Parse(conf.PathPrefix)
		if err != nil {
			return nil, fmt.Errorf("invalid redirect prefix:%s", conf.PathPrefix)
		}
		redirectPrefix = strings.TrimSuffix(u.Path, "/")
		if !strings.HasPrefix(u.Path, "/") {
			redirectPrefix = "/" + redirectPrefix
		}

	}
	r := &handler{
		WorkerBase:       drivers.Worker(id, name),
		maxRedirectCount: redirectCount,
		redirectPrefix:   redirectPrefix,
		autoRedirect:     conf.AutoRedirect,
	}

	return r, nil
}
