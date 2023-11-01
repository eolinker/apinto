package fileoutput

import (
	"github.com/eolinker/eosc"
)

type Config struct {
	Scopes []string `json:"scopes" label:"作用域"`
	File   string   `json:"file" yaml:"file" label:"文件名称"`
	Dir    string   `json:"dir" yaml:"dir" label:"文件存放目录"`
	Period string   `json:"period" yaml:"period" enum:"hour,day" label:"日志分割周期"`
	Expire int      `json:"expire" yaml:"expire" label:"日志保存时间" description:"单位：天" default:"7" minimum:"1"`
	Type   string   `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	//BodyConfig BodyConfig           `json:"body_config" yaml:"body_config" label:"请求体/响应体配置" description:"请求体/响应体配置" switch:"type===json"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

//type BodyConfig struct {
//	BodySize int    `json:"body_size" label:"请求体/响应体截取长度" description:"单位：M" default:"10" minimum:"0"`
//	BodyCode string `json:"body_code" label:"请求体/响应体编码" enum:"latin,utf8,gbk" default:"utf8"`
//}
