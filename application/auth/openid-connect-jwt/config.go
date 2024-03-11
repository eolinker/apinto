package openid_connect_jwt

import (
	"encoding/base64"

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
	Issuer                   string   `json:"issuer"`
	AuthenticatedGroupsClaim []string `json:"authenticated_groups_claim"`
}

func (u *User) Username() string {
	return base64.RawStdEncoding.EncodeToString([]byte(u.Pattern.Issuer))
}
