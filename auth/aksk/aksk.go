package aksk

import (
	"errors"
	"fmt"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"ak/sk",
	"aksk",
}

type aksk struct {
	id             string
	name           string
	hideCredential bool
	users          *akskUsers
}

func (a *aksk) Id() string {
	return a.id
}

func (a *aksk) Start() error {
	return nil
}

func (a *aksk) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}

	a.hideCredential = config.HideCredentials

	a.users = &akskUsers{
		users: config.Users,
	}

	return nil
}

func (a *aksk) Stop() error {
	return nil
}

func (a *aksk) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (a *aksk) Auth(context *http_context.Context) error {
	authorizationType, has := context.Request().Header().Get(auth.AuthorizationType)
	if !has {
		return auth.ErrorInvalidType
	}
	err := auth.CheckAuthorizationType(supportTypes, authorizationType)
	if err != nil {
		return err
	}
	//解析Authorization字符串
	encType, ak, signHeaders, signature, err := parseAuthorization(context)
	//判断配置中是否存在该ak
	for _, user := range a.users.users {
		if ak == user.AK {
			switch encType {
			case "SDK-HMAC-SHA256", "HMAC-SHA256":
				{
					//结合context内的信息与配置的sk生成新的签名，与context携带的签名进行对比
					toSign := buildToSign(context, encType, signHeaders)
					s := hmaxBySHA256(user.SK, toSign)
					if s == signature {
						// 判断鉴权是否已过期
						if user.Expire != 0 && time.Now().Unix() > user.Expire {
							return errors.New("[ak/sk_auth] authorization expired")
						}

						//若隐藏证书信息
						if a.hideCredential {
							context.ProxyRequest().Header.Del(auth.Authorization)
						}
						return nil
					}
				}
			}
		}
	}

	return errors.New("[ak/sk_auth] Invalid authorization")
}
