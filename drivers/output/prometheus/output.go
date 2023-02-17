package prometheus

import (
	"reflect"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ output.IEntryOutput = (*PromeOutput)(nil)
var _ eosc.IWorker = (*PromeOutput)(nil)

type PromeOutput struct {
	drivers.WorkerBase
	config *Config

	registry     *prometheus.Registry
	Metrics      map[string]iMetric
	MetricLabels map[string][]labelConfig
}

type labelConfig struct {
	Name  string
	Type  string
	Value string
}

func (p *PromeOutput) Output(entry eosc.IEntry) error {
	//TODO

	return nil
}

func (p *PromeOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}

	metricLabels, err := doCheck(cfg)

	//TODO 检查新旧配置的指标，若有变化，才替换Register和handler
	if reflect.DeepEqual(cfg, p.config) {
		return nil
	}

	//TODO 若path有变，更新worker路由器

	//TODO 若Scopes有变,更新scopeManager
	scopeManager.Set(p.Id(), p, cfg.Scopes)

	p.config = cfg

	return nil
}

func (p *PromeOutput) Stop() error {
	//TODO 注销指标

	//TODO 注销路由

	return nil
}

func (p *PromeOutput) Start() error {
	//TODO 注册指标

	//TODO 注册路由
	handler := promhttp.InstrumentMetricHandler(
		p.registry, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}),
	)

	scopeManager.Set(p.Id(), p, p.config.Scopes)
	return nil
}

func (p *PromeOutput) CheckSkill(skill string) bool {
	//TODO
	return true
}
