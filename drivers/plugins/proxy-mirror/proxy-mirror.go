package proxy_mirror

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ eocontext.IFilter = (*proxyMirror)(nil)
var _ http_service.HttpFilter = (*proxyMirror)(nil)

type proxyMirror struct {
	drivers.WorkerBase
	proxyConf *Config
}

func (p *proxyMirror) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {
	return http_service.DoHttpFilter(p, ctx, next)
}

func (p *proxyMirror) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	if next != nil {
		err := next.DoChain(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	//进行采样, 生成随机数判断

	//进行转发

	return nil
}

func (p *proxyMirror) Start() error {
	return nil
}

func (p *proxyMirror) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	p.proxyConf = conf

	return nil
}

func (p *proxyMirror) Stop() error {
	return nil
}

func (p *proxyMirror) Destroy() {
}

func (p *proxyMirror) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
