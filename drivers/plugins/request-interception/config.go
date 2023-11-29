package request_interception

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Config struct {
	Status      int       `json:"status" label:"响应状态码" minimum:"100" default:"200" description:"最小值：100"`
	Body        string    `json:"body" label:"响应体"`
	ContentType string    `json:"content_type" label:"响应体类型" default:"application/json" enum:"text/plain,text/html,application/json"`
	Headers     []*Header `json:"headers" label:"响应头"`
}

type Header struct {
	Key   string   `json:"key" label:"响应头Key"`
	Value []string `json:"value" label:"响应头Value"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	contentType := conf.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	return &executor{
		WorkerBase:  drivers.Worker(id, name),
		status:      conf.Status,
		body:        conf.Body,
		headers:     conf.Headers,
		contentType: conf.ContentType,
	}, nil
}
