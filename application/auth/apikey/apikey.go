package apikey

import (
	"fmt"
	"github.com/eolinker/apinto/application"
	"time"
	
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	
	"github.com/eolinker/eosc"
)

var _ application.IAuthFilter = (*apikey)(nil)

type apikey struct {
	id        string
	tokenName string
	position  string
	users     eosc.IUntyped
}

func (a *apikey) Driver() string {
	return name
}

func (a *apikey) ID() string {
	return a.ID()
}

func (a *apikey) Set(appID string, users []*application.User) {
	a.users.Set(appID, users)
}

func (a *apikey) Del(appID string) {
	a.users.Del(appID)
}

func (a *apikey) getUserMap(appID string) (map[string]*application.User, bool) {
	return nil, false
}

//Auth 鉴权处理
func (a *apikey) Auth(ctx http_service.IHttpContext) error {
	value, has := application.GetToken(a.tokenName, a.position, ctx)
	if !has {
		return fmt.Errorf("%s error: %s in %s:%s", name, application.ErrTokenNotFound, a.position, a.tokenName)
	}
	users := a.getUsers()
	for _, user := range users {
		ok := isValidUser(user.Pattern, value, user.Expire)
		if ok {
			for k, v := range user.Labels {
				ctx.SetLabel(k, v)
			}
			return nil
		}
	}
	return fmt.Errorf("%s error: %s %s", name, application.ErrInvalidToken, value)
}

func isValidUser(pattern map[string]string, value string, expire int64) bool {
	if v, ok := pattern["apikey"]; ok {
		if v == value {
			if expire == 0 || time.Now().Unix() < expire {
				return true
			}
		}
	}
	return false
}

func (a *apikey) getUsers() []*application.User {
	users := a.users.List()
	us := make([]*application.User, 0, len(users))
	for _, user := range users {
		u, ok := user.(*application.User)
		if ok {
			us = append(us, u)
		}
	}
	return us
}
