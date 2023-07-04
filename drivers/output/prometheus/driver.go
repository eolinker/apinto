package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strings"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	_, err := doCheck(v)
	return err
}

// doCheck 检查配置并返回指标的标签配置列表
func doCheck(promConf *Config) (map[string]*metricInfoCfg, error) {
	metricLabels := make(map[string]*metricInfoCfg, len(promConf.Metrics))

	if len(promConf.Metrics) == 0 {
		return nil, errorNullMetrics
	}

	tmpMetric := make(map[string]struct{}, len(promConf.Metrics))
	for _, metricConf := range promConf.Metrics {
		//格式化配置，去除空格
		metricConf.Metric = strings.TrimSpace(metricConf.Metric)
		metricConf.Collector = strings.TrimSpace(metricConf.Collector)
		metricConf.Objectives = strings.TrimSpace(metricConf.Objectives)
		metricConf.Description = strings.TrimSpace(metricConf.Description)
		formatLabels := make([]string, 0, len(metricConf.Labels))
		for _, l := range metricConf.Labels {
			formatLabels = append(formatLabels, strings.TrimSpace(l))
		}
		metricConf.Labels = formatLabels

		//指标名不能为空
		if metricConf.Metric == "" {
			return nil, errorNullMetric
		}
		//指标名查重
		if _, exist := tmpMetric[metricConf.Metric]; exist {
			return nil, fmt.Errorf(errorMetricReduplicatedFormat, metricConf.Metric)
		}
		tmpMetric[metricConf.Metric] = struct{}{}

		//校验objectives
		if metricConf.Objectives != "" {
			isMatch := utils.CheckObjectives(metricConf.Objectives)
			if !isMatch {
				return nil, fmt.Errorf(errorObjectivesFormat, metricConf.Metric, metricConf.Objectives)
			}

		} else {
			metricConf.Objectives = defaultObjectives
		}

		//校验收集类型合不合法
		if _, exist := collectorSet[metricConf.Collector]; exist {
			//指标标签不能为空
			if len(metricConf.Labels) == 0 {
				return nil, fmt.Errorf(errorNullLabelsFormat, metricConf.Metric)
			}

			labels := make([]labelConfig, 0, len(metricConf.Labels))
			tmpLabels := make(map[string]struct{}, len(metricConf.Labels))
			for _, label := range metricConf.Labels {
				cLabel, err := formatLabel(label)
				if err != nil {
					return nil, err
				}
				//标签名查重
				if _, isExist := tmpLabels[cLabel.Name]; isExist {
					return nil, fmt.Errorf(errorLabelReduplicatedFormat, metricConf.Metric, cLabel.Name)
				}
				tmpLabels[cLabel.Name] = struct{}{}

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

	scopes := make([]string, 0, len(promConf.Scopes))
	for _, scope := range promConf.Scopes {
		s := strings.TrimSpace(scope)
		if s == "" {
			return nil, errorNullScopeMetric
		}
		scopes = append(scopes, scope)
	}
	promConf.Scopes = scopes

	return metricLabels, nil
}

func formatLabel(label string) (labelConfig, error) {
	if label == "" {
		return labelConfig{}, fmt.Errorf(errorNullLabelFormat, label)
	}

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
		return labelConfig{}, fmt.Errorf(errorLabelFormat, label)
	}

	return c, nil
}

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	metricsInfo, err := doCheck(cfg)
	if err != nil {
		return nil, err
	}

	p := &PromOutput{
		WorkerBase:  drivers.Worker(id, name),
		config:      cfg,
		metricsInfo: metricsInfo,
	}

	//注册指标
	registry := prometheus.NewPedanticRegistry()

	metrics := make(map[string]iMetric, len(p.config.Metrics))
	for _, metric := range p.config.Metrics {
		m, err := newIMetric(p.metricsInfo[metric.Metric], metric.Metric, metric.Description, metric.Objectives)
		if err != nil {
			return nil, fmt.Errorf("create output %s fail: %w", p.Id(), err)
		}
		err = m.Register(registry)
		if err != nil {
			return nil, fmt.Errorf("create output %s fail: %w", p.Id(), err)
		}
		metrics[metric.Metric] = m
	}

	//注册路由
	p.registry = registry
	p.handler = promhttp.InstrumentMetricHandler(
		p.registry, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}),
	)

	//metrics路径为 /apinto/metrics/prometheus/{worker_name} 前面的/apinto/在router.SetPath里做拼接
	err = router.SetPath(p.Id(), fmt.Sprintf("/metrics/prometheus/%s", name), p)
	if err != nil {
		return nil, fmt.Errorf("create output %s fail: %w", p.Id(), err)
	}

	p.metrics = metrics

	scope_manager.Set(p.Id(), p, p.config.Scopes...)

	return p, err
}
