package aksk

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
	AK string `json:"ak"`
	SK string `json:"sk"`
}

func (u *User) Username() string {
	return u.Pattern.AK
}
