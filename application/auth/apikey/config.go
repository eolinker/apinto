package apikey

import "github.com/eolinker/apinto/application"

type Config struct {
	application.Auth
	Users []*User `json:"users"`
}

type User struct {
	application.User
	Pattern Pattern `json:"pattern"`
}

type Pattern struct {
	Apikey string `json:"apikey"`
}

func (u *User) Username() string {
	return u.Pattern.Apikey
}
