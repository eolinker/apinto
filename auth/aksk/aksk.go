package aksk

import (
	"errors"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"time"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"ak/sk",
	"aksk",
}

type aksk struct {
	id         string
	name       string
	akskConfig map[string]AKSKConfig
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

	for _, c := range config.akskConfig {
		a.akskConfig[c.AK] = c
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
	err := auth.CheckAuthorizationType(supportTypes, context.Request().Headers().Get(auth.AuthorizationType))
	if err != nil {
		return err
	}
	//解析Authorization字符串
	encType, ak, signHeaders, signature, err := parseAuthorization(context)
	//判断配置中是否存在该ak
	conf, has := a.akskConfig[ak]
	if !has {
		return errors.New("[ak/sk_auth] Invalid authorization")
	}
	// 判断鉴权是否已过期
	if conf.Expire != 0 && time.Now().Unix() > conf.Expire {
		return errors.New("[ak/sk_auth] authorization expired")
	}

	switch encType {
	case "SDK-HMAC-SHA256", "HMAC-SHA256":
		{
			//结合context内的信息与配置的sk生成新的签名，与context携带的签名进行对比
			toSign := buildToSign(context, encType, signHeaders)
			s := hmaxBySHA256(conf.SK, toSign)
			if s == signature {
				//若隐藏证书信息
				if conf.HideCredential {
					context.Proxy().DelHeader(auth.Authorization)
				}
				return nil
			}
		}
	}

	return errors.New("[ak/sk_auth] Invalid authorization")
}
