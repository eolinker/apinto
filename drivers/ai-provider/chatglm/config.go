package chatglm

import (
	"fmt"
	"net/url"

	"github.com/eolinker/eosc"
)

type Config struct {
	Base string `json:"api_base"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.Base != "" {
		u, err := url.Parse(conf.Base)
		if err != nil {
			return nil, fmt.Errorf("base url is invalid")
		}
		if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("base url is invalid")
		}
	}
	return conf, nil
}
