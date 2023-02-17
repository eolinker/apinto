package prometheus

import (
	prometheus_entry "github.com/eolinker/apinto/entries/prometheus-entry"
	"reflect"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ prometheus_entry.IOutput = (*PromOutput)(nil)
var _ eosc.IWorker = (*PromOutput)(nil)

type PromOutput struct {
	drivers.WorkerBase
	config *Config

	registry    *prometheus.Registry
	metrics     map[string]iMetric
	metricsInfo map[string]*metricInfo
}

type metricInfo struct {
	collector string
	labels    []labelConfig
}

type labelConfig struct {
	Name  string
	Type  string
	Value string
}

func (p *PromOutput) Output(metrics []string, entry prometheus_entry.IPromEntry) {
	//TODO

}

func (p *PromOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}

	metricsInfo, err := doCheck(cfg)

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

func (p *PromOutput) Stop() error {
	//TODO 注销指标

	//TODO 注销路由

	return nil
}

func (p *PromOutput) Start() error {
	//TODO 注册指标

	//TODO 注册路由
	handler := promhttp.InstrumentMetricHandler(
		p.registry, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}),
	)

	scopeManager.Set(p.Id(), p, p.config.Scopes)
	return nil
}

func (p *PromOutput) CheckSkill(skill string) bool {
	return skill == prometheus_entry.Skill
}
