package access_log

import (
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
	http_entry "github.com/eolinker/goku/http-entry"
)

type accessLog struct {
	id     string
	output []eosc.IOutput
}

func (l *accessLog) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
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

}

func (l *accessLog) Id() string {
	return l.id
}

func (l *accessLog) Start() error {
	panic("implement me")
}

func (l *accessLog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
}

func (l *accessLog) Stop() error {
	panic("implement me")
}

func (l *accessLog) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
