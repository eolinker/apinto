package fileoutput

import (
	"github.com/eolinker/eosc"
)

type Config struct {
	File      string               `json:"file" yaml:"file"`
	Dir       string               `json:"dir" yaml:"dir"`
	Period    string               `json:"period" yaml:"period"`
	Expire    int                  `json:"expire" yaml:"expire"`
	Type      string               `json:"type" yaml:"type" enum:"json,line"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}
