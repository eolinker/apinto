package auth

import "github.com/eolinker/eosc"

type Config struct {
	Auth []eosc.RequireId `json:"auth" skill:"github.com/eolinker/apinto/auth.auth.IAuth"`
}
