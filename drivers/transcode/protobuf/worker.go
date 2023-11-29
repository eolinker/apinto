package protocbuf

import (
	"errors"
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc/protoparse"

	"github.com/eolinker/apinto/drivers"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
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
	source, err := parseFiles(cfg.ProtoFiles)
	if err != nil {
		return err
	}
	w.source = source
	return nil
}

func (w *Worker) Stop() error {
	w.source = nil
	return nil
}

func (w *Worker) CheckSkill(skill string) bool {
	return grpc_descriptor.Skill == skill
}

func parseFiles(files eosc.EoFiles) (grpcurl.DescriptorSource, error) {
	descSourceFiles := map[string]string{}
	fileNames := make([]string, 0, len(files))
	for _, f := range files {
		v, err := f.DecodeData()
		if err != nil {
			return nil, fmt.Errorf("file(%s) data decode error: %v", f.Name, err)
		}
		descSourceFiles[f.Name] = string(v)
		fileNames = append(fileNames, f.Name)

	}
	p := &protoparse.Parser{Accessor: protoparse.FileContentsFromMap(descSourceFiles)}
	descSources, err := p.ParseFiles(fileNames...)
	if err != nil {
		return nil, err
	}
	return grpcurl.DescriptorSourceFromFileDescriptors(descSources...)

}
