package app

import "github.com/eolinker/apinto/application"

//Config App驱动配置
type Config struct {
	Auth    []*Auth `json:"auth"`
	Disable bool    `json:"disable"`
}

type Auth struct {
	Config    interface{}         `json:"config"`
	Type      string              `json:"type"`
	Users     []*application.User `json:"users"`
	Position  string              `json:"position"`
	TokenName string              `json:"token_name"`
}
