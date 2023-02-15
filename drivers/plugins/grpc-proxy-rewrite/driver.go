package grpc_proxy_rewrite

import (
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	conf.Authority = strings.TrimSpace(conf.Authority)

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	pw := &ProxyRewrite{
		WorkerBase:      drivers.Worker(id, name),
		host:            strings.TrimSpace(conf.Authority),
		headers:         conf.Headers,
		tls:             conf.Tls,
		skipCertificate: conf.SkipCertificate,
	}

	return pw, nil
}
