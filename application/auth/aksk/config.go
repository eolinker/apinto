package aksk

import (
	"github.com/eolinker/apinto/application"
)

type Config struct {
	application.Auth
	Users []*User `json:"users" label:"用户列表"`
}

type User struct {
	Pattern Pattern `json:"pattern" label:"用户信息"`
	application.User
}

type Pattern struct {
	AK string `json:"ak" label:"AK"`
	SK string `json:"sk" label:"SK"`
}

func (u *User) Username() string {
	return u.Pattern.AK
}
