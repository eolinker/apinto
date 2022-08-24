package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
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

func (b *basic) ID() string {
	return b.id
}

func (b *basic) Driver() string {
	return driverName
}

func (b *basic) Check(users []*application.User) error {
	us := make(map[string]*application.User)
	for _, user := range users {
		name, has := getUser(user.Pattern)
		if !has {
			return errors.New("invalid user")
		}
		_, ok := b.users.Get(name)
		if ok {
			return errors.New("user is existed")
		}
		if _, ok = us[name]; ok {
			return errors.New("user is existed")
		}
		us[name] = user
	}
	return nil
}

func (b *basic) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	if b.users == nil {
		b.users = application.NewUserManager()
	}
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			Name:           name,
			Value:          getPassword(user.Pattern),
			Expire:         user.Expire,
			Labels:         user.Labels,
			HideCredential: user.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
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

func (b *basic) Auth(ctx http_service.IHttpContext) error {
	token, has := application.GetToken(ctx, b.tokenName, b.position)
	if !has || token == "" {
		return fmt.Errorf("%s error: %s in %s:%s", driverName, application.ErrTokenNotFound, b.position, b.tokenName)
	}
	username, password := parseToken(token)
	if username == "" {
		return fmt.Errorf("%s error: %s %s", driverName, application.ErrInvalidToken, token)
	}
	user, has := b.users.Get(username)
	if has {
		if password == user.Value {
			for k, v := range user.Labels {
				ctx.SetLabel(k, v)
			}
			for k, v := range user.AppLabels {
				ctx.SetLabel(k, v)
			}
			if user.HideCredential {
				application.HideToken(ctx, b.tokenName, b.position)
			}
			return nil
		}
	}
	
	return fmt.Errorf("%s error: %s %s", driverName, application.ErrInvalidToken, token)
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
