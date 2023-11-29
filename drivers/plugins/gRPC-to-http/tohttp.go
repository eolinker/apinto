package grpc_to_http

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers"
)

type toHttp struct {
	drivers.WorkerBase
	handler eocontext.CompleteHandler
}

func (t *toHttp) DoGrpcFilter(ctx grpc_context.IGrpcContext, next eocontext.IChain) (err error) {
	if t.handler != nil {
		ctx.SetCompleteHandler(t.handler)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (t *toHttp) Start() error {
	return nil
}

func (t *toHttp) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return grpc_context.DoGrpcFilter(t, ctx, next)
}

func (t *toHttp) Destroy() {
	t.handler = nil
}

func (t *toHttp) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	descSource, err := getDescSource(string(cfg.ProtobufID))
	if err != nil {
		return err
	}
	t.handler = newComplete(descSource, cfg)
	return nil
}

func (t *toHttp) Stop() error {
	t.Destroy()
	return nil
}

func (t *toHttp) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
