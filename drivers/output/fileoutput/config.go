package fileoutput

import (
	"github.com/eolinker/eosc"
)

type Config struct {
	File      string               `json:"file" yaml:"file" label:"文件名称"`
	Dir       string               `json:"dir" yaml:"dir" label:"文件存放目录"`
	Period    string               `json:"period" yaml:"period" enum:"hour,day" label:"日志分割周期"`
	Expire    int                  `json:"expire" yaml:"expire" label:"日志保存时间" description:"单位：天" default:"7" minimum:"1"`
	Type      string               `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}
