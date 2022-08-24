package aksk

import (
	"fmt"
	"github.com/eolinker/apinto/application"
	"time"
	
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"ak/sk",
	"aksk",
}

var _ application.IAuth = (*aksk)(nil)

type aksk struct {
	id        string
	tokenName string
	position  string
	users     application.IUserManager
}

func (a *aksk) ID() string {
	return a.id
}

func (a *aksk) Driver() string {
	return a.Driver()
}

func (a *aksk) Check(appID string, users []*application.User) error {
	return a.users.Check(appID, driverName, users)
}

func (a *aksk) Set(appID string, labels map[string]string, disable bool, users []*application.User) {
	infos := make([]*application.UserInfo, 0, len(users))
	for _, user := range users {
		name, _ := getUser(user.Pattern)
		infos = append(infos, &application.UserInfo{
			AppID:          appID,
			Name:           name,
			Value:          getValue(user.Pattern),
			Expire:         user.Expire,
			Labels:         user.Labels,
			HideCredential: user.HideCredential,
			AppLabels:      labels,
			Disable:        disable,
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
	encType, ak, signHeaders, signature, err := parseAuthorization(ctx)
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
