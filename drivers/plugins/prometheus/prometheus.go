package prometheus

import (
	metric_entry "github.com/eolinker/apinto/entries/metric-entry"
	"reflect"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ eocontext.IFilter = (*prometheus)(nil)
var _ http_service.HttpFilter = (*prometheus)(nil)

type prometheus struct {
	drivers.WorkerBase
	proxy   scope_manager.IProxyOutput
	metrics []string
}

func (p *prometheus) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(p, ctx, next)
}

func (p *prometheus) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	err = next.DoChain(ctx)
	if err != nil {
		log.Error(err)
	}

	metricEntry, err := metric_entry.NewMetricEntry(ctx)
	if err != nil {
		log.Errorf("prometheus plugin id:%s DoHttpFilter fail. %w", p.Id(), err)
		return
	}

	outputs := p.proxy.List()
	for _, v := range outputs {
		o, ok := v.(metric_entry.IOutput)
		if !ok {
			log.Error("prometheus output type error,type is ", reflect.TypeOf(v))
			continue
		}
		o.Output(p.metrics, metricEntry)
	}

	return nil
}

func (p *prometheus) Destroy() {
	p.proxy = nil
}

func (p *prometheus) Start() error {
	return nil
}

func (p *prometheus) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	c, err := check(conf)
	if err != nil {
		return err
	}

	list, err := getList(c.Output)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		proxy := scope_manager.NewProxy()
		proxy.Set(list)
		p.proxy = proxy
	} else {
		p.proxy = scopeManager.Get(globalScopeName)
	}

	p.metrics = c.Metrics
	return nil
}

func (p *prometheus) Stop() error {
	return nil
}

func (p *prometheus) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
