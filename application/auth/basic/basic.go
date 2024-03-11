package basic

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/eolinker/apinto/application"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*basic)(nil)

type basic struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
}

func (b *basic) Alias() []string {
	return []string{
		"basic",
		"basic_auth",
	}
}

func (b *basic) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, b.tokenName, b.position)
	if !has || token == "" {
		return nil, false
	}
	username, password := parseToken(token)
	if username == "" {
		return nil, false
	}
	user, has := b.users.Get(username)
	if has {
		if password == user.Value {
			return user, true
		}
	}
	return nil, false
}

func (b *basic) ID() string {
	return b.id
}

func (b *basic) Driver() string {
	return driverName
}

func (b *basic) Check(appID string, users []application.ITransformConfig) error {
	us := make([]application.IUser, 0, len(users))
	for _, u := range users {
		v, ok := u.Config().(*User)
		if !ok {
			return fmt.Errorf("%s check error: invalid config type", driverName)
		}
		us = append(us, v)
	}
	return b.users.Check(appID, driverName, us)
}

func (b *basic) Set(app application.IApp, users []application.ITransformConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		v, _ := user.Config().(*User)

		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Value:          v.Pattern.Password,
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			TokenName:      b.tokenName,
			Position:       b.position,
			App:            app,
		})
	}
	b.users.Set(app.Id(), infos)
}

func (b *basic) Del(appID string) {
	b.users.DelByAppID(appID)
}

func (b *basic) UserCount() int {
	return b.users.Count()
}

func parseToken(token string) (username string, password string) {
	const basic = "basic"
	l := len(basic)

	if len(token) > l+1 && strings.ToLower(token[:l]) == basic {
		b, err := base64.StdEncoding.DecodeString(token[l+1:])
		if err != nil {
			return "", ""
		}
		cred := string(b)
		for i := 0; i < len(cred); i++ {
			if cred[i] == ':' {
				return cred[:i], cred[i+1:]
			}
		}
		return "", ""
	} else {
		return "", ""
	}
}
