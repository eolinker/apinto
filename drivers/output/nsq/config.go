package nsq

import "github.com/eolinker/eosc"

type Config struct {
	Config *NsqConf `json:"config" yaml:"config"`
}

type NsqConf struct {
	Topic      string               `json:"topic" yaml:"topic"`
	Address    string               `json:"address" yaml:"address"`
	AuthSecret string               `json:"auth_secret" yaml:"auth_secret"`
	Type       string               `json:"type" yaml:"type"`
	Formatter  eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}

func (c *NsqConf) isProducerUpdate(conf *NsqConf) bool {
	if c.Address != conf.Address || c.AuthSecret != conf.AuthSecret {
		return true
	}
	return false
}
