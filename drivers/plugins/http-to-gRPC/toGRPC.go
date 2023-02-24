package http_to_grpc

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type toGRPC struct {
	drivers.WorkerBase
	handler eocontext.CompleteHandler
}

func (t *toGRPC) Start() error {
	return nil
}

func (t *toGRPC) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(t, ctx, next)
}
func (t *toGRPC) Destroy() {
	t.handler = nil
	return
}

func (t *toGRPC) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {
	if t.handler != nil {
		ctx.SetCompleteHandler(t.handler)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (t *toGRPC) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	descSource, err := getDescSource(string(cfg.ProtobufID), cfg.Reflect)
	if err != nil {
		return err
	}
	t.handler = newComplete(descSource, cfg)
	return nil
}

func (t *toGRPC) Stop() error {
	return nil
}

func (t *toGRPC) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
