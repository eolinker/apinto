package http_to_grpc

import (
	"fmt"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	once.Do(func() {
		bean.Autowired(&worker)
	})

	descSource, err := getDescSource(string(conf.ProtobufID), conf.Reflect)
	if err != nil {
		return nil, err
	}
	return &toGRPC{
		WorkerBase: drivers.Worker(id, name),
		handler:    newComplete(descSource, conf),
	}, nil
}

func getDescSource(protobufID string, reflect bool) (grpc_descriptor.IDescriptor, error) {
	if reflect {
		return nil, nil
	}
	if protobufID == "" {
		return nil, fmt.Errorf("protobuf id is empty")
	}
	if worker == nil {
		return nil, fmt.Errorf("protobuf worker is not initialized")
	}
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
