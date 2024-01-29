package openid_connect_jwt

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ohler55/ojg/jp"

	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/application"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ application.IAuth = (*jwt)(nil)

type jwt struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
	cache     scope_manager.IProxyOutput[resources.ICache]
	once      sync.Once
}

func (o *jwt) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, o.tokenName, o.position)
	if !has || token == "" {
		return nil, false
	}
	id, obj, has := verify(token)
	if !has {
		return nil, false
	}
	info, has := o.users.Get(id)
	if !has {
		return nil, false
	}
	exprs, ok := info.Additional.([]jp.Expr)
	if !ok {
		return nil, false
	}
	result := make([]interface{}, 0, len(exprs))
	for _, expr := range exprs {
		v := expr.Get(obj)
		if len(v) > 0 {
			result = append(result, v...)
		}
	}
	ctx.WithValue("acl_groups", result)
	return info, true
}

func (o *jwt) ID() string {
	return o.id
}

func (o *jwt) Driver() string {
	return driverName
}

func (o *jwt) Check(appID string, users []application.ITransformConfig) error {
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

func (o *jwt) Set(app application.IApp, users []application.ITransformConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	idMap := make(map[string]struct{})
	oldIDMap := manager.GetIssuerIDMap(app.Id())
	for _, user := range users {
		v, _ := user.Config().(*User)
		exprs := make([]jp.Expr, 0, len(v.Pattern.AuthenticatedGroupsClaim))
		for _, key := range v.Pattern.AuthenticatedGroupsClaim {
			if !strings.HasPrefix(key, "$.") {
				key = fmt.Sprintf("$.%s", key)
			}
			expr, err := jp.ParseString(key)
			if err != nil {
				log.Errorf("parse key %w, key: %s", err, key)
				continue
			}
			exprs = append(exprs, expr)
		}
		manager.Set(v.Username(), &IssuerConfig{
			ID:     v.Username(),
			Issuer: v.Pattern.Issuer,
		})
		delete(oldIDMap, v.Username())
		idMap[v.Username()] = struct{}{}
		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Value:          strings.Join(v.Pattern.AuthenticatedGroupsClaim, ","),
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			TokenName:      o.tokenName,
			Position:       o.position,
			App:            app,
			Additional:     exprs,
		})
	}
	for id := range oldIDMap {
		manager.Del(id)
	}
	manager.SetIssuerIDMap(app.Id(), idMap)
	o.users.Set(app.Id(), infos)
}

func (o *jwt) Del(appID string) {
	o.users.DelByAppID(appID)
	oldIDMap, _ := manager.DelIssuerIDMap(appID)
	for id := range oldIDMap {
		manager.Del(id)
	}
}

func (o *jwt) UserCount() int {
	return o.users.Count()
}
