package dubbo2_to_http

import (
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/entries/ctx_key"
	"github.com/eolinker/apinto/entries/router"
)

var _ eocontext.IFilter = (*ToHttp)(nil)
var _ dubbo2_context.DubboFilter = (*ToHttp)(nil)

type ToHttp struct {
	drivers.WorkerBase
	method      string
	path        string
	contentType string
	params      []param
}

func (t *ToHttp) DoDubboFilter(ctx dubbo2_context.IDubbo2Context, next eocontext.IChain) (err error) {

	retryValue := ctx.Value(ctx_key.CtxKeyRetry)
	retry, ok := retryValue.(int)
	if !ok {
		retry = router.DefaultRetry
	}

	timeoutValue := ctx.Value(ctx_key.CtxKeyTimeout)
	timeout, ok := timeoutValue.(time.Duration)
	if !ok {
		timeout = router.DefaultTimeout
	}

	complete := NewComplete(retry, timeout, t.contentType, t.path, t.method, t.params)

	ctx.SetCompleteHandler(complete)

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (t *ToHttp) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return dubbo2_context.DoDubboFilter(t, ctx, next)
}

func (t *ToHttp) Destroy() {

}

func (t *ToHttp) Start() error {
	return nil
}

func (t *ToHttp) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	t.path = conf.Path
	t.method = conf.Method
	t.contentType = conf.ContentType

	params := make([]param, 0, len(conf.Params))

	for _, val := range conf.Params {
		params = append(params, param{
			className: val.ClassName,
			fieldName: val.FieldName,
		})
	}
	t.params = params
	return nil
}

func (t *ToHttp) Stop() error {
	return nil
}

func (t *ToHttp) CheckSkill(skill string) bool {
	return dubbo2_context.FilterSkillName == skill
}

type param struct {
	className string
	fieldName string
}
