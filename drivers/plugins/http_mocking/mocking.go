package http_mocking

import (
	"encoding/json"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	log "github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
)

var _ eocontext.IFilter = (*Mocking)(nil)
var _ http_context.HttpFilter = (*Mocking)(nil)

type Mocking struct {
	drivers.WorkerBase
	responseStatus  int
	contentType     string
	responseExample string
	responseSchema  string
	responseHeader  map[string]string
	handler         eocontext.CompleteHandler
}

func (m *Mocking) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {
	if m.handler != nil {
		ctx.SetCompleteHandler(m.handler)
	}

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (m *Mocking) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(m, ctx, next)
}

func (m *Mocking) Destroy() {
	return
}

func (m *Mocking) Start() error {
	return nil
}

func (m *Mocking) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	m.responseSchema = conf.ResponseSchema
	m.responseExample = conf.ResponseExample
	m.contentType = conf.ContentType
	m.responseStatus = conf.ResponseStatus
	m.responseHeader = conf.ResponseHeader

	jsonSchema := make(map[string]interface{})

	if conf.ResponseSchema != "" {
		if err = json.Unmarshal([]byte(conf.ResponseSchema), &jsonSchema); err != nil {
			log.Errorf("create mocking err=%s,jsonSchema=%s", err.Error(), conf.ResponseSchema)
			return err
		}
	}

	m.handler = NewComplete(m.responseStatus, m.contentType, m.responseExample, jsonSchema, m.responseHeader)

	return nil
}

func (m *Mocking) Stop() error {
	return nil
}

func (m *Mocking) CheckSkill(skill string) bool {
	return true
}
