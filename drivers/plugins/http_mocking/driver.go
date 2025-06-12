package http_mocking

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/log"
	"strings"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	if conf.ResponseStatus < 100 {
		conf.ResponseStatus = 200
	}

	if conf.ContentType == contentTypeJson {
		if len(strings.TrimSpace(conf.ResponseSchema)) == 0 && len(strings.TrimSpace(conf.ResponseExample)) == 0 {
			log.Errorf("mocking check schema is null && example is null ")
			return nil, errors.New("param err")
		}
		if len(strings.TrimSpace(conf.ResponseExample)) > 0 {
			var val interface{}

			if err = json.Unmarshal([]byte(conf.ResponseExample), &val); err != nil {
				log.Errorf("mocking check example Format  err = %s  example=%s", err.Error(), conf.ResponseExample)
				return nil, errors.New("param err")
			}

		}
		if len(strings.TrimSpace(conf.ResponseSchema)) > 0 {
			var val interface{}

			if err = json.Unmarshal([]byte(conf.ResponseSchema), &val); err != nil {
				log.Errorf("mocking check Schema Format  err = %s  Schema=%s", err.Error(), conf.ResponseSchema)
				return nil, errors.New("param err")
			}

		}
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	once.Do(func() {
		bean.Autowired(&worker)
	})

	jsonSchema := make(map[string]interface{})

	if conf.ResponseSchema != "" {
		if err := json.Unmarshal([]byte(conf.ResponseSchema), &jsonSchema); err != nil {
			log.Errorf("create mocking err=%s,jsonSchema=%s", err.Error(), conf.ResponseSchema)
			return nil, err
		}
	}

	return &Mocking{
		WorkerBase: drivers.Worker(id, name),
		handler:    NewComplete(conf.ResponseStatus, conf.ContentType, conf.ResponseExample, jsonSchema, conf.ResponseHeader),
	}, nil
}

func getDescSource(protobufID string) (grpc_descriptor.IDescriptor, error) {

	w, ok := worker.Get(protobufID)
	if ok {
		v, vOk := w.(grpc_descriptor.IDescriptor)
		if !vOk {
			return nil, fmt.Errorf("invalid protobuf id: %s", protobufID)
		}
		return v, nil
	}
	return nil, fmt.Errorf("protobuf worker(%s) is not exist", protobufID)

}
