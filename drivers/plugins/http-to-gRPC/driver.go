package http_to_grpc

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc"
	"github.com/fullstorydev/grpcurl"
)

func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	descSource, err := getDescSource(conf.ProtobufID, workers)
	if err != nil {
		return nil, err
	}
	return &toGRPC{
		WorkerBase: drivers.Worker(id, name),
		handler:    newComplete(descSource, conf),
	}, nil
}

func getDescSource(protobufID eosc.RequireId, workers map[eosc.RequireId]eosc.IWorker) (grpcurl.DescriptorSource, error) {
	var descSource grpcurl.DescriptorSource
	if protobufID != "" {
		worker, ok := workers[protobufID]
		if ok {
			v, ok := worker.(grpc_descriptor.IDescriptor)
			if !ok {
				return nil, fmt.Errorf("invalid protobuf id: %s", protobufID)
			}
			descSource = v.Descriptor()
		}
	}
	return descSource, nil
}
