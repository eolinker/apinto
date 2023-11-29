package grpc_to_http

import (
	"fmt"

	"github.com/eolinker/eosc/common/bean"

	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	if conf.ProtobufID == "" {
		return nil, fmt.Errorf("protobuf id is empty")
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	once.Do(func() {
		bean.Autowired(&worker)
	})
	descSource, err := getDescSource(string(conf.ProtobufID))
	if err != nil {
		return nil, err
	}
	return &toHttp{
		WorkerBase: drivers.Worker(id, name),
		handler:    newComplete(descSource, conf),
	}, nil
}

func getDescSource(protobufID string) (grpc_descriptor.IDescriptor, error) {

	w, ok := worker.Get(protobufID)
	if ok {
		v, ok := w.(grpc_descriptor.IDescriptor)
		if !ok {
			return nil, fmt.Errorf("invalid protobuf id: %s", protobufID)
		}
		return v, nil
	}
	return nil, fmt.Errorf("protobuf worker(%s) is not exist", protobufID)

}
