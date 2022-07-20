package access_log

import (
	http_entry "github.com/eolinker/apinto/http-entry"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/context"
	http_service "github.com/eolinker/eosc/context/http-context"
	"github.com/eolinker/eosc/log"
)

var _ context.IFilter = (*accessLog)(nil)
var _ http_service.HttpFilter = (*accessLog)(nil)

type accessLog struct {
	*Driver
	id     string
	output []output.IEntryOutput
}

func (l *accessLog) DoFilter(ctx context.Context, next context.IChain) (err error) {
	return http_service.DoHttpFilter(l, ctx, next)
}

func (l *accessLog) DoHttpFilter(ctx http_service.IHttpContext, next context.IChain) (err error) {
	err = next.DoChain(ctx)
	if err != nil {
		log.Error(err)
	}
	entry := http_entry.NewEntry(ctx)
	for _, o := range l.output {
		err = o.Output(entry)
		if err != nil {
			log.Error("access log http-entry error:", err)
			continue
		}
	}
	return nil
}

func (l *accessLog) Destroy() {
	l.output = nil
}

func (l *accessLog) Id() string {
	return l.id
}

func (l *accessLog) Start() error {
	return nil
}

func (l *accessLog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	c, err := l.check(conf)
	if err != nil {
		return err
	}
	list, err := l.getList(c.Output)
	if err != nil {
		return err
	}

	l.output = list
	return nil
}

func (l *accessLog) Stop() error {
	return nil
}

func (l *accessLog) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
