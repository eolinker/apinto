package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return doCheck(v)
}

func doCheck(v *Config) error {
	promConf := v

	if match := utils.CheckUrlPath(promConf.Path); !match {
		return fmt.Errorf(errorPathFormat, promConf.Path)
	}

	if len(promConf.Metrics) == 0 {
		return errorNullMetrics
	}

	//先校验指标，再校验指标的标签
	for _, metricConf := range promConf.Metrics {
		if metricType, exist := metricSet[metricConf.Metric]; exist {
			if len(metricConf.Labels) == 0 {
				return fmt.Errorf(errorNullLabelsFormat, metricConf.Metric)
			}

			for _, label := range metricConf.Labels {
				if _, has := metricLabelSet[metricType][label]; !has {
					return fmt.Errorf(errorLabelFormat, label)
				}
			}
		} else {
			return fmt.Errorf(errorMetricFormat, metricConf.Metric)
		}
	}

	//TODO 对标签进行排序

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

	worker := &PromeOutput{
		WorkerBase: drivers.Worker(id, name),
		config:     cfg,
	}
	return worker, err
}
