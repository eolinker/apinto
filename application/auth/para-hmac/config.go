package para_hmac

import "github.com/eolinker/apinto/application"

type Config struct {
	application.Auth
	Users []*User `json:"users" label:"用户列表"`
}

type User struct {
	Pattern Pattern `json:"pattern" label:"用户信息"`
	application.User
}

type Pattern struct {
	AppID  string `json:"app_id"`
	AppKey string `json:"app_key"`
}

func (u *User) Username() string {
	return u.Pattern.AppID
}
