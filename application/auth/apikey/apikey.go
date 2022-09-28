package apikey

import (
	"fmt"

	"github.com/eolinker/apinto/application"
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
	return a.id
}

func (a *apikey) Check(appID string, users []application.ITransformConfig) error {
	us := make([]application.IUser, 0, len(users))
	for _, u := range users {
		v, ok := u.Config().(*User)
		if !ok {
			return fmt.Errorf("%s check error: invalid config type", driverName)
		}
		us = append(us, v)
	}
	return a.users.Check(appID, driverName, us)
}

func (a *apikey) Set(app application.IApp, users []application.ITransformConfig) {

	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		v, _ := user.Config().(*User)

		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Value:          v.Username(),
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			TokenName:      a.tokenName,
			Position:       a.position,
			App:            app,
		})
	}
	a.users.Set(app.Id(), infos)
}

func (a *apikey) Del(appID string) {
	a.users.DelByAppID(appID)
}

//GetUser 鉴权处理
func (a *apikey) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, a.tokenName, a.position)
	if !has || token == "" {
		return nil, false
	}
	user, has := a.users.Get(token)
	if has {
		return user, true
	}
	return nil, false
}

func (a *apikey) Driver() string {
	return driverName
}

func (a *apikey) UserCount() int {
	return a.users.Count()
}
