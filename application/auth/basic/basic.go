package basic

import (
	"encoding/base64"
	"github.com/eolinker/apinto/application"
	"strings"

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
		return nil, true
	}
	user, has := b.users.Get(username)
	if has {
		if password == user.Value {
			return user, true
		}
		return nil, true
	}
	return nil, false
}

func (b *basic) ID() string {
	return b.id
}

func (b *basic) Driver() string {
	return driverName
}

func (b *basic) Check(appID string, users []*application.User) error {
	return b.users.Check(appID, driverName, users)
}

func (b *basic) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			AppID:          appID,
			Name:           name,
			Value:          getPassword(user.Pattern),
			Expire:         user.Expire,
			Labels:         user.Labels,
			HideCredential: user.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
			TokenName:      b.tokenName,
			Position:       b.position,
		})
	}
	b.users.Set(appID, infos)
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

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["username"]; ok {
		return v, true
	}
	return "", false
}

func getPassword(pattern map[string]string) string {
	if v, ok := pattern["password"]; ok {
		return v
	}
	return ""
}
