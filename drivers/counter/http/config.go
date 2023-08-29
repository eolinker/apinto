package http

import (
	"fmt"
	"net/url"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Scopes           []string          `json:"scopes" label:"作用域"`
	URI              string            `json:"uri" label:"地址"`
	Method           string            `json:"method" label:"请求方式" enum:"GET,POST,PUT,PATCH"`
	ContentType      string            `json:"content_type" label:"请求体类型" enum:"json,form-data"`
	Headers          map[string]string `json:"headers" label:"请求头"`
	QueryParam       map[string]string `json:"query_param" label:"请求参数"`
	BodyParam        map[string]string `json:"body_param" label:"请求体参数"`
	ResponseJsonPath string            `json:"response_json_path" label:"响应体JsonPath路径"`
}

func (c *Config) Validate() error {
	u, err := url.Parse(c.URI)
	if err != nil {
		return fmt.Errorf("parse uri error:%w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("uri scheme must be http or https")
	}
	if c.Method == "" {
		c.Method = "GET"
	}
	if c.Method != "GET" && c.Method != "POST" && c.Method != "PUT" && c.Method != "PATCH" {
		return fmt.Errorf("method must be GET,POST,PUT,PATCH")
	}
	if c.ContentType == "" {
		c.ContentType = "json"
	}

	return nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := conf.Validate()
	if err != nil {
		return nil, err
	}
	bc := &Executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = bc.reset(conf)
	if err != nil {
		return nil, err
	}
	return bc, nil
}
