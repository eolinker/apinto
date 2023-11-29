package jwt

import (
	"fmt"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/application"
)

var _ application.IAuth = (*jwt)(nil)

type jwt struct {
	id        string
	tokenName string
	position  string
	cfg       *Rule
	users     application.IUserManager
}

func (j *jwt) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, j.tokenName, j.position)
	if !has || token == "" {
		return nil, false
	}
	name, err := j.doJWTAuthentication(token)
	if err != nil {
		log.DebugF("[%s] get user error:%s", driverName, token)
		return nil, false
	}
	return j.users.Get(name)
}

func (j *jwt) ID() string {
	return j.id
}

func (j *jwt) Driver() string {
	return driverName
}

func (j *jwt) Check(appID string, users []application.ITransformConfig) error {
	us := make([]application.IUser, 0, len(users))
	for _, u := range users {
		v, ok := u.Config().(*User)
		if !ok {
			return fmt.Errorf("%s check error: invalid config type", driverName)
		}
		us = append(us, v)
	}
	return j.users.Check(appID, driverName, us)
}

func (j *jwt) Set(app application.IApp, users []application.ITransformConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		v, _ := user.Config().(*User)
		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			TokenName:      j.tokenName,
			Position:       j.position,
			App:            app,
		})
	}
	j.users.Set(app.Id(), infos)
}

func (j *jwt) Del(appID string) {
	j.users.DelByAppID(appID)
}

func (j *jwt) UserCount() int {
	return j.users.Count()
}
