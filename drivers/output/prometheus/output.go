package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/entries/metric-entry"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ metric_entry.IOutput = (*PromOutput)(nil)
var _ eosc.IWorker = (*PromOutput)(nil)

type PromOutput struct {
	drivers.WorkerBase
	config *Config

	registry    *prometheus.Registry
	metrics     map[string]iMetric
	metricsInfo map[string]*metricInfoCfg
}

type metricInfoCfg struct {
	collector string
	labels    []labelConfig
}

type labelConfig struct {
	Name  string
	Type  string
	Value string
}

func (p *PromOutput) Output(metrics []string, entry eosc.IMetricEntry) {
	proxyEntries := entry.Children("proxies")

	for _, metric := range metrics {
		//若prometheus插件的metric 在output中不存在则跳过
		metricInfo, exist := p.metricsInfo[metric]
		if !exist {
			continue
		}

		//判断是请求collector还是转发collector
		switch collectorSet[metricInfo.collector] {
		case typeRequestMetric:
			p.writeMetric(p.metrics[metric], metricInfo, []eosc.IMetricEntry{entry})
		case typeProxyMetric:
			p.writeMetric(p.metrics[metric], metricInfo, proxyEntries)

		}

	}

}

func (p *PromOutput) writeMetric(metric iMetric, metricInfo *metricInfoCfg, entries []eosc.IMetricEntry) {
	for _, entry := range entries {
		labels := make(map[string]string, len(metricInfo.labels))
		for _, l := range metricInfo.labels {
			switch l.Type {
			case labelTypeVar:
				labels[l.Name] = entry.Read(l.Value)
			case labelTypeConst:
				//常量标签
				labels[l.Name] = l.Value
			}

		}
		value, _ := entry.GetFloat(metricInfo.collector)
		metric.Observe(value, labels)

	}
}

func (p *PromOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}

	metricsInfo, err := doCheck(cfg)

	registry := prometheus.NewPedanticRegistry()
	metrics := make(map[string]iMetric, len(p.config.Metrics))
	for _, metric := range cfg.Metrics {
		m, err := newIMetric(metric.Collector, metric.Metric, metric.Description, metricsInfo[metric.Metric], metric.Objectives)
		if err != nil {
			return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
		}
		err = m.Register(registry)
		if err != nil {
			return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
		}
		metrics[metric.Metric] = m
	}

	//注册路由
	handler := promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
	router.DeletePath(p.Id())
	err = router.AddPath(p.Id(), p.config.Path, handler)
	if err != nil {
		return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
	}

	p.metricsInfo = metricsInfo
	p.registry = registry
	p.metrics = metrics
	p.config = cfg

	//TODO 可优化点
	//检查新旧配置的指标，若有变化，才替换Register和handler
	//if reflect.DeepEqual(cfg, p.config) {
	//	return nil
	//}
	//若path有变，更新worker路由器
	//若Scopes有变,更新scopeManager
	scopeManager.Set(p.Id(), p, cfg.Scopes)

	return nil
}

func (p *PromOutput) Stop() error {
	//注销路由
	router.DeletePath(p.Id())

	return nil
}

func (p *PromOutput) Start() error {
	//注册指标
	registry := prometheus.NewPedanticRegistry()

	metrics := make(map[string]iMetric, len(p.config.Metrics))
	for _, metric := range p.config.Metrics {
		m, err := newIMetric(metric.Collector, metric.Metric, metric.Description, p.metricsInfo[metric.Metric], metric.Objectives)
		if err != nil {
			return fmt.Errorf("start output %s fail: %w", p.Id(), err)
		}
		err = m.Register(registry)
		if err != nil {
			return fmt.Errorf("start output %s fail: %w", p.Id(), err)
		}
		metrics[metric.Metric] = m
	}

	//注册路由
	handler := promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)

	err := router.AddPath(p.Id(), p.config.Path, handler)
	if err != nil {
		return fmt.Errorf("start output %s fail: %w", p.Id(), err)
	}

	p.registry = registry
	p.metrics = metrics

	scopeManager.Set(p.Id(), p, p.config.Scopes)
	return nil
}

func (p *PromOutput) CheckSkill(skill string) bool {
	return skill == metric_entry.Skill
}
