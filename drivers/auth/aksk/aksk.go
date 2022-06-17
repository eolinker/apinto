package aksk

import (
	"errors"
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"time"

	"github.com/eolinker/apinto/auth"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"ak/sk",
	"aksk",
}

type aksk struct {
	id             string
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
	c, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}

	a.hideCredential = c.HideCredentials

	a.users = &akskUsers{
		users: c.Users,
	}

	return nil
}

func (a *aksk) Stop() error {
	return nil
}

func (a *aksk) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (a *aksk) Auth(context http_service.IHttpContext) error {
	authorizationType := context.Request().Header().GetHeader(auth.AuthorizationType)
	if authorizationType == "" {
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
							context.Proxy().Header().DelHeader(auth.Authorization)
						}
						return nil
					}
				}
			}
		}
	}

	return errors.New("[ak/sk_auth] Invalid authorization")
}
