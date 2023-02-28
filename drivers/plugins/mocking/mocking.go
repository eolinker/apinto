package mocking

import (
	"encoding/json"
	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	log "github.com/eolinker/eosc/log"
)

var _ eocontext.IFilter = (*Mocking)(nil)

type Mocking struct {
	drivers.WorkerBase
	responseStatus  int
	contentType     string
	responseExample string
	responseSchema  string
	handler         eocontext.CompleteHandler
}

func (t *Mocking) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {

	if t.handler != nil {
		ctx.SetCompleteHandler(t.handler)
	}

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (t *Mocking) Destroy() {
	return
}

func (t *Mocking) Start() error {
	return nil
}

func (t *Mocking) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	t.responseSchema = conf.ResponseSchema
	t.responseExample = conf.ResponseExample
	t.contentType = conf.ContentType
	t.responseStatus = conf.ResponseStatus

	var descriptor grpc_descriptor.IDescriptor
	if t.contentType == contentTypeGrpc {
		descriptor, err = getDescSource(string(conf.ProtobufID))
		if err != nil {
			return err
		}
	}

	jsonSchema := make(map[string]interface{})

	if conf.ResponseSchema != "" {
		if err = json.Unmarshal([]byte(conf.ResponseSchema), &jsonSchema); err != nil {
			log.Errorf("create mocking err=%s,jsonSchema=%s", err.Error(), conf.ResponseSchema)
			return err
		}
	}

	t.handler = NewComplete(t.responseStatus, t.contentType, t.responseExample, jsonSchema, descriptor)

	return nil
}

func (t *Mocking) Stop() error {
	return nil
}

func (t *Mocking) CheckSkill(skill string) bool {
	return true
}
