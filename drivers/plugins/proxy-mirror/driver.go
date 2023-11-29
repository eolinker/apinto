package proxy_mirror

import (
	"time"

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

	pm := &proxyMirror{
		WorkerBase:  drivers.Worker(id, name),
		randomRange: conf.SampleConf.RandomRange,
		randomPivot: conf.SampleConf.RandomPivot,
		service:     newMirrorService(conf.Addr, conf.PassHost, conf.Host, time.Duration(conf.Timeout)),
		conf:        conf,
	}

	return pm, nil
}
