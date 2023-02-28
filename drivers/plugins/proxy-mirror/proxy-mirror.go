package proxy_mirror

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"math/rand"
	"time"
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
	if p.proxyConf != nil {
		//进行采样, 生成随机数判断
		rand.Seed(time.Now().UnixNano())
		randomNum := rand.Intn(p.proxyConf.SampleConf.RandomRange + 1) //[0,range]范围内整型
		if randomNum <= p.proxyConf.SampleConf.RandomPivot {           //若随机数在[0,pivot]范围内则进行转发
			//进行转发
			go sendMirrorProxy(p.proxyConf, ctx)
		}
	}

	return nil
}

func sendMirrorProxy(proxyCfg *Config, ctx http_service.IHttpContext) {
	//先判断当前Ctx是否能Copy,若可以就进行copy并且设置新的APP

	//send

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
	p.proxyConf = nil
}

func (p *proxyMirror) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
