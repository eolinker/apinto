package access_log

import (
	"reflect"

	"github.com/eolinker/apinto/drivers"
	scope_manager "github.com/eolinker/apinto/drivers/scope-manager"
	http_entry "github.com/eolinker/apinto/http-entry"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ eocontext.IFilter = (*accessLog)(nil)
var _ http_service.HttpFilter = (*accessLog)(nil)

type accessLog struct {
	drivers.WorkerBase
	proxy scope_manager.IProxyOutput
}

func (l *accessLog) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(l, ctx, next)
}

func (l *accessLog) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	err = next.DoChain(ctx)
	if err != nil {
		log.Error(err)
	}
	entry := http_entry.NewEntry(ctx)

	outputs := l.proxy.List()
	for _, v := range outputs {
		o, ok := v.(output.IEntryOutput)
		if !ok {
			log.Error("access log output type error,type is ", reflect.TypeOf(v))
			continue
		}
		err = o.Output(entry)
		if err != nil {
			log.Error("access log http-entry error:", err)
			continue
		}
	}

	return nil
}

func (l *accessLog) Destroy() {
	l.proxy = nil
}

func (l *accessLog) Start() error {
	return nil
}

func (l *accessLog) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
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

	} else {
		l.proxy = scopeManager.Get("access_log")
	}

	return nil
}

func (l *accessLog) Stop() error {
	return nil
}

func (l *accessLog) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
