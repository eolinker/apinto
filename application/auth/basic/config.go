package basic

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
	Username string `json:"username" label:"用户名"`
	Password string `json:"password" label:"密码"`
}

func (u *User) Username() string {
	return u.Pattern.Username
}
