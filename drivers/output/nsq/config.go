package nsq

import "github.com/eolinker/eosc"

type Config struct {
	Topic      string                 `json:"topic" yaml:"topic"`
	Address    []string               `json:"address" yaml:"address"`
	ClientConf map[string]interface{} `json:"nsq_conf" yaml:"nsq_conf"`
	Type       string                 `json:"type" yaml:"type" enum:"json,line"`
	Formatter  eosc.FormatterConfig   `json:"formatter" yaml:"formatter"`
}
