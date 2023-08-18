package body_check

import (
	"errors"
	"net/http"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*BodyCheck)(nil)
var _ eocontext.IFilter = (*BodyCheck)(nil)

type BodyCheck struct {
	drivers.WorkerBase
	isEmpty            bool
	allowedPayloadSize int
}

func (b *BodyCheck) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(b, ctx, next)
}

func (b *BodyCheck) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	if ctx.Request().Method() == http.MethodPost || ctx.Request().Method() == http.MethodPut || ctx.Request().Method() == http.MethodPatch {
		body, err := ctx.Request().Body().RawBody()
		if err != nil {
			return err
		}
		bodySize := len([]rune(string(body)))
		if !b.isEmpty && bodySize < 1 {
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetBody([]byte("Body is required"))
			return errors.New("Body is required")
		}
		if b.allowedPayloadSize > 0 && bodySize > b.allowedPayloadSize {
			ctx.Response().SetStatus(413, "413")
			ctx.Response().SetBody([]byte("The request body is too large"))
			return errors.New("The request entity is too large")
		}
	}

	return next.DoChain(ctx)
}

func (b *BodyCheck) Start() error {
	return nil
}

func (b *BodyCheck) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return errors.New("invalid config")
	}
	b.isEmpty = cfg.IsEmpty
	b.allowedPayloadSize = cfg.AllowedPayloadSize * 1024
	return nil
}

func (b *BodyCheck) Stop() error {
	return nil
}

func (b *BodyCheck) Destroy() {
	return
}

func (b *BodyCheck) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
