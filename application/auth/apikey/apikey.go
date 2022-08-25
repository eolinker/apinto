package apikey

import (
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

func (a *apikey) Check(appID string, users []*application.User) error {
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
			TokenName:      a.tokenName,
			Position:       a.position,
		})
	}
	a.users.Set(appID, infos)
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

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["apikey"]; ok {
		return v, true
	}
	return "", false
}
