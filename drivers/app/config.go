package app

import (
	"encoding/json"
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/application/auth"
)

//Config App驱动配置
type Config struct {
	Auth    []*Auth           `json:"auth"`
	Labels  map[string]string `json:"labels"`
	Disable bool              `json:"disable"`
}

type Auth struct {
	Type      string                    `json:"type"`
	Users     []*application.BaseConfig `json:"users"`
	Position  string                    `json:"position"`
	TokenName string                    `json:"token_name"`
	Config    *application.BaseConfig   `json:"config"`
}

func (a *Auth) Reset(originVal reflect.Value, targetVal reflect.Value, variables eosc.IVariable) ([]string, error) {
	bytes, err := json.Marshal(originVal.Interface())
	if err != nil {
		log.Error("auth config unmarshal error: ", err)
		return nil, err
	}
	var tmp Auth
	err = json.Unmarshal(bytes, &tmp)
	if err != nil {
		return nil, err
	}
	log.Debug("set type: ", string(bytes))

	f, err := auth.GetFactory(tmp.Type)
	if err != nil {
		return nil, err
	}
	if tmp.Config != nil && f.ConfigType() != nil {
		err = tmp.Config.SetType(f.ConfigType())
		if err != nil {
			return nil, err
		}
	}
	if f.UserType() != nil {
		for _, user := range tmp.Users {
			err = user.SetType(f.UserType())
			if err != nil {
				return nil, err
			}
		}
	}
	targetVal.Set(reflect.ValueOf(tmp))
	return nil, nil
}
