package app

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

var (
	extenders eosc.IExtenderDrivers
	ones      sync.Once
)

type Config struct {
	Auth []*AuthConfig `json:"auth" label:"鉴权列表"`
}

type AuthConfig struct {
	Config    interface{} `json:"config"`
	DriverID  string      `json:"driver_id"`
	Position  string      `json:"position"`
	TokenName string      `json:"token_name"`
	Users     []*User     `json:"users"`
}

type User struct {
	Expire         int               `json:"expire"`
	Labels         map[string]string `json:"labels"`
	Pattern        string            `json:"pattern"`
	HideCredential bool              `json:"hide_credential"`
}

func Check(v interface{}) (*Config, error) {
	ones.Do(func() {
		bean.Autowired(&extenders)
	})
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	for _, a := range cfg.Auth {
		fac, has := extenders.GetDriver(a.DriverID)
		if !has {
			return nil, eosc.ErrorDriverNotExist
		}
		driver, err := fac.Create("auth", "", "", "", nil)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
