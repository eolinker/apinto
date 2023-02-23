package protocbuf

import (
	"errors"

	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc"
	"github.com/fullstorydev/grpcurl"
)

type Worker struct {
	drivers.WorkerBase
	source grpcurl.DescriptorSource
}

func (w *Worker) Descriptor() grpcurl.DescriptorSource {
	return w.source
}

func (w *Worker) Start() error {
	return nil
}

func (w *Worker) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return errors.New("illegal config type")
	}
	parseFiles(cfg.ProtoFiles)
	return nil
}

func (w *Worker) Stop() error {
	w.source = nil
	return nil
}

func (w *Worker) CheckSkill(skill string) bool {
	return grpc_descriptor.Skill == skill
}

func parseFiles(files eosc.EoFiles) grpcurl.DescriptorSource {
	descSourceFiles := map[string]string{}
	for _, f := range files {
		descSourceFiles[f.Name] = f.Data
	}
	return nil
}
