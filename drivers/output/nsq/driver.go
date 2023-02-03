package nsq

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

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
func doCheck(conf *Config) error {

	nsqConf := conf
	if nsqConf == nil {
		return errNsqConfNull
	}
	if nsqConf.Topic == "" {
		return errTopicNull
	}
	if len(nsqConf.Address) == 0 {
		return errAddressNull
	}
	if nsqConf.Type == "" {
		nsqConf.Type = "line"
	}
	switch nsqConf.Type {
	case "line", "json":
	default:
		return errFormatterType
	}

	if len(nsqConf.Formatter) == 0 {
		return errFormatterConf
	}

	return nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := doCheck(conf)
	if err != nil {
		return nil, err
	}
	worker := &NsqOutput{
		WorkerBase: drivers.Worker(id, name),
	}
	worker.config = conf
	return worker, nil

}
