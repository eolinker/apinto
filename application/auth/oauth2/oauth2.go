package oauth2

import (
	"fmt"
	"sync"

	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/application"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*oauth2)(nil)

type oauth2 struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
	cache     scope_manager.IProxyOutput[resources.ICache]
	once      sync.Once
}

func (o *oauth2) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, o.tokenName, o.position)
	if !has || token == "" {
		return nil, false
	}
	o.once.Do(func() {
		o.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	list := o.cache.List()
	if len(list) < 1 {
		return nil, false
	}
	clientID, err := validToken(ctx.Context(), list[0], token)
	if err != nil {
		log.Error("valid token error:", err, "token:", token)
		return nil, false
	}

	return o.users.Get(clientID)
}

func (o *oauth2) ID() string {
	return o.id
}

func (o *oauth2) Driver() string {
	return driverName
}

func (o *oauth2) Check(appID string, users []application.ITransformConfig) error {
	us := make([]application.IUser, 0, len(users))
	for _, u := range users {
		v, ok := u.Config().(*User)
		if !ok {
			return fmt.Errorf("%s check error: invalid config type", driverName)
		}
		us = append(us, v)
	}
	return o.users.Check(appID, driverName, us)
}

func (o *oauth2) Set(app application.IApp, users []application.ITransformConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		v, _ := user.Config().(*User)
		client := &Client{
			Pattern: &v.Pattern,
			Expire:  v.Expire,
		}
		if v.Pattern.HashSecret {
			hr, err := extractHashRule(v.Pattern.ClientSecret)
			if err != nil {
				log.Error("extract hash error:", err, "client secret:", v.Pattern.ClientSecret)
				continue
			}
			log.DebugF("hash rule: ", *hr)
			client.hashRule = hr
		}
		registerClient(v.Pattern.ClientId, client)

		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Value:          v.Pattern.ClientSecret,
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			TokenName:      o.tokenName,
			Position:       o.position,
			App:            app,
		})
	}
	o.users.Set(app.Id(), infos)
}

func (o *oauth2) Del(appID string) {
	o.users.DelByAppID(appID)
}

func (o *oauth2) UserCount() int {
	return o.users.Count()
}
