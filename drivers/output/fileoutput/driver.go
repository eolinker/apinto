package fileoutput

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return doCheck(v)
}
func doCheck(v *Config) error {
	fileConf := v
	if fileConf == nil {
		return errorNilConfig
	}

	if fileConf.Dir == "" {
		return errorConfDir
	}
	if fileConf.File == "" {
		return errorConfFile
	}
	if fileConf.Period != "day" && fileConf.Period != "hour" {
		return errorConfPeriod
	}

	if fileConf.Expire == 0 {
		fileConf.Expire = 3
	}
	if fileConf.Type == "" {
		fileConf.Type = "line"
	}

	if len(fileConf.Formatter) == 0 {
		return errFormatterConf
	}
	return nil
}
func check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, errorConfigType
	}
	err := doCheck(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil

}

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := doCheck(cfg)
	if err != nil {
		return nil, err
	}

	worker := &FileOutput{
		WorkerBase: drivers.Worker(id, name),
		config:     cfg,
	}
	return worker, err
}
