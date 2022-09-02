package jwt

import (
	"fmt"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
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

func (j *jwt) Check(appID string, users []*application.BaseConfig) error {
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

func (j *jwt) Set(appID string, labels map[string]string, disable bool, users []*application.BaseConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		v, _ := user.Config().(*User)
		infos = append(infos, &application.UserInfo{
			AppID:          appID,
			Name:           v.Username(),
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
			TokenName:      j.tokenName,
			Position:       j.position,
		})
	}
	j.users.Set(appID, infos)
}

func (j *jwt) Del(appID string) {
	j.users.DelByAppID(appID)
}

func (j *jwt) UserCount() int {
	return j.users.Count()
}

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["username"]; ok {
		return v, true
	}
	return "", false
}
