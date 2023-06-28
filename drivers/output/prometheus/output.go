package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var _ output.IMetrics = (*PromOutput)(nil)
var _ eosc.IWorker = (*PromOutput)(nil)

type PromOutput struct {
	drivers.WorkerBase
	config *Config

	registry    *prometheus.Registry
	handler     http.Handler
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

func (p *PromOutput) Collect(metrics []string, entry eosc.IMetricEntry) {
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

func (p *PromOutput) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	p.handler.ServeHTTP(writer, request)
}

func (p *PromOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}

	metricsInfo, err := doCheck(cfg)

	//若指标配置有变动，更新metrics和registry的配置
	allMetrics, isMetricsUpdate, err := p.FormatUpdateMetrics(p.config.Metrics, cfg.Metrics, metricsInfo)
	if err != nil {
		return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
	}
	var newRegistry *prometheus.Registry
	var handler http.Handler
	if isMetricsUpdate {
		newRegistry = prometheus.NewPedanticRegistry()
		for _, m := range allMetrics {
			err = m.Register(newRegistry)
			if err != nil {
				return fmt.Errorf("reset output %s register metric fail: %w", p.Id(), err)
			}
		}
		handler = promhttp.InstrumentMetricHandler(
			newRegistry, promhttp.HandlerFor(newRegistry, promhttp.HandlerOpts{}),
		)
	}

	if isMetricsUpdate {
		p.registry = newRegistry
		p.handler = handler
	}

	//若Scopes有变,更新scopeManager
	if checkScopesChange(p.config.Scopes, cfg.Scopes) {
		scope_manager.Set(p.Id(), p, cfg.Scopes...)
	}

	p.metricsInfo = metricsInfo
	p.metrics = allMetrics
	p.config = cfg

	return nil
}

// FormatUpdateMetrics 返回更新后的metrics配置
func (p *PromOutput) FormatUpdateMetrics(oldMetrics, newMetrics []*MetricConfig, newMetricsInfo map[string]*metricInfoCfg) (map[string]iMetric, bool, error) {
	allMetrics := make(map[string]iMetric, len(newMetrics))
	isMetricsUpdate := false

	oldMetricsMap := make(map[string]*MetricConfig)
	newMetricsMap := make(map[string]*MetricConfig)
	for _, mc := range oldMetrics {
		oldMetricsMap[mc.Metric] = mc
	}
	for _, mc := range newMetrics {
		newMetricsMap[mc.Metric] = mc
	}

	//对比新旧metric配置，若某个旧metric配置有改动，则New一个新的IMetric；若没有改变，则继续使用旧的IMetric
	for m, oldMC := range oldMetricsMap {
		if newMC, exist := newMetricsMap[m]; !exist {
			isMetricsUpdate = true
		} else {
			if checkMetricConfigChange(oldMC, newMC) {
				isMetricsUpdate = true
				newIMC, err := newIMetric(newMetricsInfo[newMC.Metric], newMC.Metric, newMC.Description, newMC.Objectives)
				if err != nil {
					return nil, false, err
				}
				allMetrics[m] = newIMC
			} else {
				allMetrics[m] = p.metrics[m]
			}

		}
	}

	for m, newMC := range newMetricsMap {
		if _, exist := oldMetricsMap[m]; !exist {
			newIMC, err := newIMetric(newMetricsInfo[newMC.Metric], newMC.Metric, newMC.Description, newMC.Objectives)
			if err != nil {
				return nil, false, err
			}
			isMetricsUpdate = true
			allMetrics[m] = newIMC
		}
	}

	return allMetrics, isMetricsUpdate, nil
}

func (p *PromOutput) Stop() error {
	//注销路由
	router.DeletePath(p.Id())
	p.registry = nil
	p.metrics = nil
	p.metricsInfo = nil
	scope_manager.Del(p.Id())
	return nil
}

func (p *PromOutput) Start() error {

	return nil
}

func (p *PromOutput) CheckSkill(skill string) bool {
	return skill == output.Skill
}
