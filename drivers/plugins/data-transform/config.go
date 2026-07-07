package data_transform

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	RequestTransform       bool              `json:"request_transform" label:"请求转换"`
	ResponseTransform      bool              `json:"response_transform" label:"响应转换"`
	TargetRequestXMLType   string            `json:"target_request_xml_type" label:"请求转换目标XML类型" enum:"application/xml,text/xml" default:"application/xml"`
	TargetRequestJsonType  string            `json:"target_request_json_type" label:"请求转换目标JSON类型" enum:"application/json" default:"application/json"`
	TargetResponseXMLType  string            `json:"target_response_xml_type" label:"响应转换目标XML类型" enum:"application/xml,text/xml" default:"application/xml"`
	TargetResponseJsonType string            `json:"target_response_json_type" label:"响应转换目标JSON类型" enum:"application/json" default:"application/json"`
	XMLRootTag             string            `json:"xml_root_tag" label:"XML根标签"`
	XMLDeclaration         map[string]string `json:"xml_declaration" label:"XML声明"`
	ErrorType              string            `json:"error_type" label:"报错数据类型" default:"json" enum:"json,xml"`
}

func check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if conf.XMLDeclaration == nil {
		conf.XMLDeclaration = map[string]string{
			"version":  "1.0",
			"encoding": "UTF-8",
		}
	}
	if conf.XMLRootTag == "" {
		conf.XMLRootTag = "root"
	}
	if conf.TargetRequestJsonType == "" {
		conf.TargetRequestJsonType = "application/json"
	}
	if conf.TargetRequestXMLType == "" {
		conf.TargetRequestXMLType = "application/xml"
	}
	if conf.TargetResponseJsonType == "" {
		conf.TargetResponseJsonType = "application/json"
	}
	if conf.TargetResponseXMLType == "" {
		conf.TargetResponseXMLType = "application/xml"
	}
	return nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	bc := &executor{
		WorkerBase: drivers.Worker(id, name),
		conf:       conf,
	}

	return bc, nil
}
