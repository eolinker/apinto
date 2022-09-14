package aksk

import (
	"fmt"
	"time"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ application.IAuth = (*aksk)(nil)

type aksk struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
}

func (a *aksk) GetUser(ctx http_service.IHttpContext) (*application.UserInfo, bool) {
	token, has := application.GetToken(ctx, a.tokenName, a.position)
	if !has || token == "" {
		return nil, false
	}
	//解析Authorization字符串
	encType, ak, signHeaders, signature, err := parseAuthorization(token)
	if err != nil {
		log.DebugF("[%s] get user error: %s", driverName, err)
		return nil, true
	}
	user, has := a.users.Get(ak)
	if has {
		switch encType {
		case "SDK-HMAC-SHA256", "HMAC-SHA256":
			{
				//结合context内的信息与配置的sk生成新的签名，与context携带的签名进行对比
				toSign := buildToSign(ctx, encType, signHeaders)
				s := hMaxBySHA256(user.Value, toSign)
				if s == signature {
					return user, true
				}
			}
		default:
			return nil, true
		}
	}
	return nil, false
}

func (a *aksk) ID() string {
	return a.id
}

func (a *aksk) Driver() string {
	return driverName
}

func (a *aksk) Check(appID string, users []*application.BaseConfig) error {
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

func (a *aksk) Set(appID string, labels map[string]string, disable bool, users []*application.BaseConfig) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, u := range users {
		v, _ := u.Config().(*User)

		infos = append(infos, &application.UserInfo{
			AppID:          appID,
			Name:           v.Username(),
			Value:          v.Pattern.SK,
			Expire:         v.Expire,
			Labels:         v.Labels,
			HideCredential: v.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
			TokenName:      a.tokenName,
			Position:       a.position,
		})
	}
	a.users.Set(appID, infos)
}

func (a *aksk) Del(appID string) {
	a.users.DelByAppID(appID)
}

func (a *aksk) UserCount() int {
	return a.users.Count()
}

func (a *aksk) Auth(ctx http_service.IHttpContext) error {
	token, has := application.GetToken(ctx, a.tokenName, a.position)
	if !has || token == "" {
		return fmt.Errorf("%s error: %s in %s:%s", driverName, application.ErrTokenNotFound, a.position, a.tokenName)
	}
	//解析Authorization字符串
	encType, ak, signHeaders, signature, err := parseAuthorization(token)
	if err != nil {
		return fmt.Errorf("%s error: %s", driverName, err.Error())
	}
	user, has := a.users.Get(ak)
	if has {
		switch encType {
		case "SDK-HMAC-SHA256", "HMAC-SHA256":
			{
				//结合context内的信息与配置的sk生成新的签名，与context携带的签名进行对比
				toSign := buildToSign(ctx, encType, signHeaders)
				s := hMaxBySHA256(user.Value, toSign)
				if s == signature {
					// 判断鉴权是否已过期
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
			}
		}
	}

	return fmt.Errorf("%s error: %s %s", driverName, application.ErrInvalidToken, token)
}

func getUser(pattern map[string]string) (string, bool) {
	if v, ok := pattern["ak"]; ok {
		return v, true
	}
	return "", false
}

func getValue(pattern map[string]string) string {
	if v, ok := pattern["sk"]; ok {
		return v
	}
	return ""
}
