package proxy_mirror

import (
	"math/rand"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
)

var _ eocontext.IFilter = (*proxyMirror)(nil)

type proxyMirror struct {
	drivers.WorkerBase
	randomRange int
	randomPivot int
	service     *mirrorService
	conf        *Config
}

func (p *proxyMirror) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {
	//进行采样, 生成随机数判断
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(p.randomRange + 1) //[0,range]范围内整型
	if randomNum <= p.randomPivot {           //若随机数在[0,pivot]范围内则进行转发
		setMirrorProxy(p.service, ctx)
	}

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func setMirrorProxy(service *mirrorService, ctx eocontext.EoContext) {
	//先判断当前Ctx是否能Copy
	if !ctx.IsCloneable() {
		log.Info(errUnsupportedContextType)
		return
	}
	//给ctx设置新的FinishHandler
	newCompleteHandler, err := newMirrorHandler(ctx, service)
	if err != nil {
		log.Info(err)
		return
	}
	ctx.SetCompleteHandler(newCompleteHandler)
}

func (p *proxyMirror) Start() error {
	if p.service == nil {
		p.service = newMirrorService(p.conf.Addr, p.conf.PassHost, p.conf.Host, time.Duration(p.conf.Timeout))
	}
	return nil
}

func (p *proxyMirror) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}

	oldService := p.service
	p.service = newMirrorService(conf.Addr, conf.PassHost, conf.Host, time.Duration(conf.Timeout))
	p.randomRange = conf.SampleConf.RandomRange
	p.randomPivot = conf.SampleConf.RandomPivot
	p.conf = conf
	if oldService != nil {
		oldService.stop()
	}
	return nil
}

func (p *proxyMirror) Stop() error {
	if p.service != nil {
		p.service.stop()
	}
	p.service = nil
	return nil
}

func (p *proxyMirror) Destroy() {
	if p.service != nil {
		p.service.stop()
	}
	p.service = nil
	return
}

func (p *proxyMirror) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
