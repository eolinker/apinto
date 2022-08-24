package apikey

import (
	"fmt"
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"
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
	return a.id
}

func (a *apikey) Check(appID string, users []*application.User) error {
	log.Debug("check appID:", appID)
	return a.users.Check(appID, driverName, users)
}

func (a *apikey) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			AppID:          appID,
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
