package response_filter

import (
	"fmt"
	"github.com/eolinker/eosc"
)

type Config struct {
	BodyFilter       []string `json:"body_filter" label:"响应体过滤字段"`
	HeaderFilter     []string `json:"header_filter" label:"响应头过滤字段"`
	HeaderFilterType string   `json:"header_filter_type" label:"响应头过滤类型" enum:"black,white" default:"black"`
	BodyFilterType   string   `json:"body_filter_type" label:"响应体过滤类型" enum:"black,white" default:"black"`
}

func check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if conf.HeaderFilterType == "" {
		conf.HeaderFilterType = "black"
	}
	if conf.HeaderFilterType != "white" && conf.HeaderFilterType != "black" {
		return fmt.Errorf("header_filter_type must be white or black")
	}
	if conf.BodyFilterType == "" {
		conf.BodyFilterType = "black"
	}
	if conf.BodyFilterType != "white" && conf.BodyFilterType != "black" {
		return fmt.Errorf("body_filter_type must be white or black")
	}
	return nil
}
