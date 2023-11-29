package response_file_parse

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "response_file_parse"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	largeWarnText := "large"
	if conf.LargeWarnText != "" {
		largeWarnText = conf.LargeWarnText
	}
	validSuffix := make(map[string]struct{})
	for key := range defaultValidSuf {
		validSuffix[key] = struct{}{}
	}
	for _, s := range conf.FileSuffix {
		validSuffix[s] = struct{}{}
	}
	return &executor{
		WorkerBase:   drivers.Worker(id, name),
		fileKey:      conf.FileKey,
		validSuf:     validSuffix,
		largeWarn:    conf.LargeWarn << 20,
		largeWarnStr: largeWarnText,
	}, nil
}
