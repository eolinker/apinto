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
	Labels     map[string]string `json:"labels" label:"应用标签"`
	Disable    bool              `json:"disable" label:"是否禁用"`
	Additional []*Additional     `json:"additional" label:"额外参数"`
	Auth       []*Auth           `json:"auth" label:"鉴权列表" eotype:"interface"`
}

type Auth struct {
	Type      string                    `json:"type" label:"鉴权类型"`
	TokenName string                    `json:"token_name" label:"token名称"`
	Position  string                    `json:"position" label:"token位置" enum:"header,query,body"`
	Config    *application.BaseConfig   `json:"config" label:"配置信息" eotype:"object"`
	Users     []*application.BaseConfig `json:"users" label:"用户列表"`
}

type Additional struct {
	Key      string            `json:"key" label:"参数名"`
	Value    string            `json:"value" label:"参数值"`
	Position string            `json:"position" label:"参数位置" enum:"header,query,body"`
	Conflict string            `json:"conflict" label:"参数存在替换规则" enum:"convert,origin,error"`
	Labels   map[string]string `json:"labels"`
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
