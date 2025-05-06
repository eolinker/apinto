package monitor

import (
	monitor_entry "github.com/eolinker/apinto/entries/monitor-entry"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ eocontext.IFilter = (*worker)(nil)
var _ http_service.HttpFilter = (*worker)(nil)

type worker struct {
	drivers.WorkerBase
}

func (l *worker) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(l, ctx, next)
}

func (l *worker) bodyFinish(ctx http_service.IHttpContext) {
	points := monitor_entry.ReadProxy(ctx)
	points = append(points, monitor_entry.ReadRequest(ctx)...)
	monitorManager.Output(l.Id(), points)
	return
}

func (l *worker) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	ctx.Proxy().AppendBodyFinish(l.bodyFinish)
	log.Debug("start monitor...")
	apiID := ctx.GetLabel("api_id")
	monitorManager.ConcurrencyAdd(apiID, 1)
	err = next.DoChain(ctx)
	if err != nil {
		log.Error(err)
	}

	if ctx.Response().IsBodyStream() {
		return nil
	}
	l.bodyFinish(ctx)
	return nil
}

func (l *worker) Destroy() {
}

func (l *worker) Start() error {
	return nil
}

func (l *worker) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	return nil
}

func (l *worker) Stop() error {

	return nil
}

func (l *worker) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
