package grpc_to_http

import (
	"fmt"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "grpc_to_http"
)

var (
	once   = sync.Once{}
	worker eosc.IWorkers
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}

func Check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.ProtobufID == "" {
		return fmt.Errorf("protobuf id is empty")
	}

	return nil
}
