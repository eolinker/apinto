package jwt

import (
	"fmt"
	"github.com/eolinker/apinto/application"
	"time"
	
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*jwt)(nil)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"jwt",
}

type jwt struct {
	id        string
	tokenName string
	position  string
	cfg       *Config
	users     application.IUserManager
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

func (j *jwt) Auth(ctx http_service.IHttpContext) error {
	token, has := application.GetToken(ctx, j.tokenName, j.position)
	if !has || token == "" {
		return fmt.Errorf("%s error: %s in %s:%s", driverName, application.ErrTokenNotFound, j.position, j.tokenName)
	}
	
	name, err := j.doJWTAuthentication(token)
	if err != nil {
		return err
	}
	user, has := j.users.Get(name)
	if has {
		if user.Expire <= time.Now().Unix() && user.Expire != 0 {
			return fmt.Errorf("%s error: %s", driverName, application.ErrTokenExpired)
		}
		for k, v := range user.Labels {
			ctx.SetLabel(k, v)
		}
		for k, v := range user.AppLabels {
			ctx.SetLabel(k, v)
		}
		if user.HideCredential {
			application.HideToken(ctx, j.tokenName, j.position)
		}
		return nil
	}
	
	return fmt.Errorf("%s error: %s %s", driverName, application.ErrInvalidToken, token)
}

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["username"]; ok {
		return v, true
	}
	return "", false
}
