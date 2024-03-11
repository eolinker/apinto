package data_transform

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	RequestTransform  bool              `json:"request_transform" label:"请求转换"`
	ResponseTransform bool              `json:"response_transform" label:"响应转换"`
	XMLRootTag        string            `json:"xml_root_tag" label:"XML根标签"`
	XMLDeclaration    map[string]string `json:"xml_declaration" label:"XML声明"`
	ErrorType         string            `json:"error_type" label:"报错数据类型" default:"json" enum:"json,xml"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	bc := &executor{
		WorkerBase: drivers.Worker(id, name),
		conf:       conf,
	}

	return bc, nil
}
