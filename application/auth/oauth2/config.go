package oauth2

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
	ClientId     string   `json:"client_id" label:"客户端ID"`
	ClientSecret string   `json:"client_secret" label:"客户端密钥"`
	ClientType   string   `json:"client_type" label:"客户端类型" enum:"public,confidential"`
	HashSecret   bool     `json:"hash_secret" label:"是否Hash加密"`
	RedirectUrls []string `json:"redirect_urls" label:"重定向URL列表"`
}

func (u *User) Username() string {
	return u.Pattern.ClientId
}
