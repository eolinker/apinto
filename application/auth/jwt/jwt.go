package jwt

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*jwt)(nil)

type jwt struct {
	id        string
	tokenName string
	position  string
	cfg       *Config
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

func (j *jwt) Check(appID string, users []*application.User) error {
	return j.users.Check(appID, driverName, users)
}

func (j *jwt) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			AppID:          appID,
			Name:           name,
			Expire:         user.Expire,
			Labels:         user.Labels,
			HideCredential: user.HideCredential,
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
