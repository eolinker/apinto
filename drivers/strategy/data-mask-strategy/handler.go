package data_mask_strategy

import (
	"fmt"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type handler struct {
	name          string
	filter        strategy.IFilter
	priority      int
	maskExecutors []mask.IMaskDriver
}

func newHandler(conf *Config) (*handler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}
	maskExecutors := make([]mask.IMaskDriver, 0, len(conf.DataMask.Rules))
	for _, rule := range conf.DataMask.Rules {
		maskFunc, err := mask.GenMaskFunc(rule.Mask)
		if err != nil {
			return nil, err
		}
		fac, has := mask.GetMaskFactory(rule.Match.Type)
		if !has {
			return nil, fmt.Errorf("match type not found: %s", rule.Match.Type)
		}
		e, err := fac.Create(rule, maskFunc)
		if err != nil {
			return nil, err
		}
		maskExecutors = append(maskExecutors, e)
	}
	return &handler{
		name:          conf.Name,
		filter:        filter,
		priority:      conf.Priority,
		maskExecutors: maskExecutors,
	}, nil
}

func (h *handler) RequestExec(ctx eocontext.EoContext) error {

	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpCtx.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	contentType := httpCtx.Proxy().ContentType()

	for _, e := range h.maskExecutors {
		body, err = e.Exec(body)
		if err != nil {
			log.Errorf("request mask exec error: (%v),handler name: (%s),rule: (%s)", err, h.name, e.String())
			continue
		}
	}
	httpCtx.Proxy().Body().SetRaw(contentType, body)

	return nil
}

func (h *handler) ResponseExec(ctx eocontext.EoContext) error {
	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}

	body := httpCtx.Response().GetBody()
	if len(body) < 1 {
		return nil
	}

	for _, e := range h.maskExecutors {
		body, err = e.Exec(body)
		if err != nil {
			log.Errorf("response mask exec error: (%v),handler name: (%s),rule: (%s)", err, h.name, e.String())
			continue
		}
	}
	httpCtx.Response().SetBody(body)
	return nil
}
