package apikey

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/application"
	"time"
	
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*apikey)(nil)

type apikey struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
}

func (a *apikey) ID() string {
	return a.ID()
}

func (a *apikey) Check(users []*application.User) error {
	us := make(map[string]*application.User)
	for _, user := range users {
		name, has := getUser(user.Pattern)
		if !has {
			return errors.New("invalid user")
		}
		_, ok := a.users.Get(name)
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

func (a *apikey) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	if a.users == nil {
		a.users = application.NewUserManager()
	}
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			Name:           name,
			Value:          name,
			Expire:         user.Expire,
			Labels:         user.Labels,
			HideCredential: user.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
		})
	}
	a.users.Set(appID, infos)
}

func (a *apikey) Del(appID string) {
	a.users.DelByAppID(appID)
}

//Auth 鉴权处理
func (a *apikey) Auth(ctx http_service.IHttpContext) error {
	token, has := application.GetToken(ctx, a.tokenName, a.position)
	if !has {
		return fmt.Errorf("%s error: %s in %s:%s", driverName, application.ErrTokenNotFound, a.position, a.tokenName)
	}
	user, has := a.users.Get(token)
	if has {
		if user.Expire <= time.Now().Unix() {
			return fmt.Errorf("%s error: %s", driverName, application.ErrTokenExpired)
		}
		for k, v := range user.Labels {
			ctx.SetLabel(k, v)
		}
		for k, v := range user.AppLabels {
			ctx.SetLabel(k, v)
		}
		if user.HideCredential {
			application.HideToken(ctx, a.tokenName, a.position)
		}
		return nil
	}
	
	return fmt.Errorf("%s error: %s %s", driverName, application.ErrInvalidToken, token)
}

func (a *apikey) Driver() string {
	return driverName
}

func (a *apikey) UserCount() int {
	return a.users.Count()
}

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["apikey"]; ok {
		return v, true
	}
	return "", false
}
