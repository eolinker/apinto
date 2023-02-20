package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
	"strings"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	_, err := doCheck(v)
	return err
}

// doCheck 检查配置并返回指标的标签配置列表
func doCheck(promConf *Config) (map[string]*metricInfoCfg, error) {
	metricLabels := make(map[string]*metricInfoCfg, len(promConf.Metrics))

	if match := utils.CheckUrlPath(promConf.Path); !match {
		return nil, fmt.Errorf(errorPathFormat, promConf.Path)
	}

	if len(promConf.Metrics) == 0 {
		return nil, errorNullMetrics
	}

	tmpMetric := make(map[string]struct{}, len(promConf.Metrics))
	for _, metricConf := range promConf.Metrics {
		//指标名不能为空
		if metricConf.Metric == "" {
			return nil, errorNullMetric
		}
		//指标名查重
		if _, exist := tmpMetric[metricConf.Metric]; exist {
			return nil, fmt.Errorf(errorMetricReduplicatedFormat, metricConf.Metric)
		}
		tmpMetric[metricConf.Metric] = struct{}{}

		//校验收集类型合不合法
		if _, exist := collectorSet[metricConf.Collector]; exist {
			//指标标签不能为空
			if len(metricConf.Labels) == 0 {
				return nil, fmt.Errorf(errorNullLabelsFormat, metricConf.Metric)
			}

			labels := make([]labelConfig, 0, len(metricConf.Labels))
			tmpLabels := make(map[string]struct{}, len(metricConf.Labels))
			for _, label := range metricConf.Labels {
				//标签名查重
				if _, isExist := tmpLabels[metricConf.Metric]; isExist {
					return nil, fmt.Errorf(errorLabelReduplicatedFormat, metricConf.Metric, label)
				}
				tmpLabels[label] = struct{}{}

				cLabel, err := formatLabel(label)
				if err != nil {
					return nil, err
				}
				labels = append(labels, cLabel)
			}

			metricLabels[metricConf.Metric] = &metricInfoCfg{
				collector: metricConf.Collector,
				labels:    labels,
			}
		} else {
			return nil, fmt.Errorf(errorCollectorFormat, metricConf.Metric)
		}

	}

	//TODO 对标签进行排序

	return metricLabels, nil
}

func formatLabel(labelExp string) (labelConfig, error) {
	label := strings.TrimSpace(labelExp)

	c := labelConfig{
		Name:  "",
		Type:  labelTypeConst,
		Value: "",
	}

	if strings.HasPrefix(label, "$") {
		c.Type = labelTypeVar
		label = label[1:]
	}

	asIdx := strings.Index(label, " as ")
	if asIdx != -1 {
		c.Name = label[asIdx+4:]
		c.Value = label[:asIdx]
	} else {
		c.Name = label
		c.Value = label
	}
	if c.Name == "" || c.Value == "" {
		return labelConfig{}, fmt.Errorf(errorLabelFormat, labelExp)
	}

	return c, nil
}

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	metricsInfo, err := doCheck(cfg)
	if err != nil {
		return nil, err
	}

	worker := &PromOutput{
		WorkerBase:  drivers.Worker(id, name),
		config:      cfg,
		metricsInfo: metricsInfo,
	}
	return worker, err
}
