package prometheus

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/entries/metric-entry"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var _ metric_entry.IOutput = (*PromOutput)(nil)
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

func (p *PromOutput) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	p.handler.ServeHTTP(writer, request)
}

func (p *PromOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}

	metricsInfo, err := doCheck(cfg)

	//若path有变，更新router
	if checkPathChange(p.config.Path, cfg.Path) {
		//重新设置路由
		err = router.SetPath(p.Id(), cfg.Path, p)
		if err != nil {
			return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
		}
	}

	//若指标配置有变动，更新metrics和registry的配置
	toDelMetrics, toAddMetrics, allMetrics, err := p.FormatUpdateMetrics(p.config.Metrics, cfg.Metrics, metricsInfo)
	if err != nil {
		return fmt.Errorf("reset output %s fail: %w", p.Id(), err)
	}
	for _, toDel := range toDelMetrics {
		toDel.UnRegister(p.registry)
	}

	for _, toAdd := range toAddMetrics {
		err = toAdd.Register(p.registry)
		if err != nil {
			return fmt.Errorf("reset output %s register metric fail: %w", p.Id(), err)
		}
	}

	//若Scopes有变,更新scopeManager
	if checkScopesChange(p.config.Scopes, cfg.Scopes) {
		scopeManager.Set(p.Id(), p, cfg.Scopes)
	}

	p.metricsInfo = metricsInfo
	p.metrics = allMetrics
	p.config = cfg

	return nil
}

// FormatUpdateMetrics 分别返回待注销的metric，待注册的metric，以及更新后的metric配置
func (p *PromOutput) FormatUpdateMetrics(oldMetrics, newMetrics []*MetricConfig, newMetricsInfo map[string]*metricInfoCfg) ([]iMetric, []iMetric, map[string]iMetric, error) {
	toDeleteIMetrics := make([]iMetric, 0, len(oldMetrics))
	toAddIMetrics := make([]iMetric, 0, len(newMetrics))
	allMetrics := make(map[string]iMetric, len(newMetrics))

	oldMetricsMap := make(map[string]*MetricConfig)
	newMetricsMap := make(map[string]*MetricConfig)
	for _, mc := range oldMetrics {
		oldMetricsMap[mc.Metric] = mc
	}
	for _, mc := range newMetrics {
		newMetricsMap[mc.Metric] = mc
	}

	/*找出待注销的metric,和待注册的metric.
	若旧metric在新metric配置中不存在则注销;
	若旧metric在新metric配置中存在且配置不一致，则将旧metric注销，并且注册新的同名metric
	*/
	for m, oldMC := range oldMetricsMap {
		if newMC, exist := newMetricsMap[m]; !exist {
			toDeleteIMetrics = append(toDeleteIMetrics, p.metrics[m])
		} else {
			if checkMetricConfigChange(oldMC, newMC) {
				toDeleteIMetrics = append(toDeleteIMetrics, p.metrics[m])
				newIMC, err := newIMetric(newMetricsInfo[newMC.Metric], newMC.Metric, newMC.Description, newMC.Objectives)
				if err != nil {
					return nil, nil, nil, err
				}
				toAddIMetrics = append(toAddIMetrics, newIMC)
				allMetrics[m] = newIMC
			} else {
				allMetrics[m] = p.metrics[m]
			}

		}
	}

	// 找出新metric配置中待注册的metric
	for m, newMC := range newMetricsMap {
		if _, exist := oldMetricsMap[m]; !exist {
			newIMC, err := newIMetric(newMetricsInfo[newMC.Metric], newMC.Metric, newMC.Description, newMC.Objectives)
			if err != nil {
				return nil, nil, nil, err
			}
			toAddIMetrics = append(toAddIMetrics, newIMC)
			allMetrics[m] = newIMC
		}
	}

	return toDeleteIMetrics, toAddIMetrics, allMetrics, nil
}

func (p *PromOutput) Stop() error {
	//注销路由
	router.DeletePath(p.Id())
	p.registry = nil
	p.metrics = nil
	p.metricsInfo = nil

	return nil
}

func (p *PromOutput) Start() error {
	//注册指标
	registry := prometheus.NewPedanticRegistry()

	metrics := make(map[string]iMetric, len(p.config.Metrics))
	for _, metric := range p.config.Metrics {
		m, err := newIMetric(p.metricsInfo[metric.Metric], metric.Metric, metric.Description, metric.Objectives)
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
	p.handler = promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)

	err := router.SetPath(p.Id(), p.config.Path, p)
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
