package para_hmac

import (
	"net/url"

	"github.com/eolinker/apinto/application"
)

import (
	"fmt"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*executor)(nil)

type executor struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
}

func (a *executor) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, a.tokenName, a.position)
	if !has || token == "" {
		return nil, false
	}
	appId := ctx.Request().Header().GetHeader("X-App-Id")
	if appId == "" {
		return nil, false
	}
	user, has := a.users.Get(appId)
	if !has {
		return nil, false
	}
	sequenceNo := ctx.Request().Header().GetHeader("X-Sequence-No")
	timestamp := ctx.Request().Header().GetHeader("X-Timestamp")
	body, _ := ctx.Request().Body().RawBody()
	signText := ctx.Request().Header().GetHeader("X-Signature")
	verifySign := sign(appId, user.Value, timestamp, sequenceNo, string(body))
	escapeSign := url.QueryEscape(verifySign)
	if verifySign == signText || escapeSign == signText {
		return user, true
	}

	return nil, false
}

func (a *executor) ID() string {
	return a.id
}

func (a *executor) Driver() string {
	return driverName
}

func (a *executor) Check(appID string, users []application.ITransformConfig) error {
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

func (a *executor) Set(app application.IApp, users []application.ITransformConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, u := range users {
		v, _ := u.Config().(*User)

		infos = append(infos, &application.UserInfo{
			Name:           v.Username(),
			Value:          v.Pattern.AppKey,
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			App:            app,
			TokenName:      a.tokenName,
			Position:       a.position,
		})
	}
	a.users.Set(app.Id(), infos)
}

func (a *executor) Del(appID string) {
	a.users.DelByAppID(appID)
}

func (a *executor) UserCount() int {
	return a.users.Count()
}
