package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"strings"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/auth"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"basic",
	"basic_auth",
	"basic-auth",
	"basicauth",
}

type basic struct {
	id             string
	hideCredential bool
	users          *basicUsers
}

type basicUsers struct {
	users []User
}

func (b *basicUsers) check(ctx http_service.IHttpContext, username string, password string) error {
	for _, u := range b.users {
		if u.Username == username && u.Password == password {
			if u.Expire == 0 || time.Now().Unix() < u.Expire {
				//将label set进context
				for k, v := range u.Labels {
					ctx.SetLabel(k, v)
				}

				return nil
			}
			return auth.ErrorExpireUser
		}
	}
	return auth.ErrorInvalidUser
}

func (b *basic) Id() string {
	return b.id
}

func (b *basic) Start() error {
	return nil
}

func (b *basic) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	b.users = &basicUsers{
		cfg.User,
	}
	b.hideCredential = cfg.HideCredentials
	return nil
}

func (b *basic) Stop() error {
	return nil
}

func (b *basic) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (b *basic) Auth(ctx http_service.IHttpContext) error {
	authorizationType := ctx.Request().Header().GetHeader(auth.AuthorizationType)
	if authorizationType == "" {
		return auth.ErrorInvalidType
	}
	err := auth.CheckAuthorizationType(supportTypes, authorizationType)
	if err != nil {
		return err
	}
	authorization := ctx.Request().Header().GetHeader(auth.Authorization)
	if b.hideCredential {
		ctx.Proxy().Header().DelHeader(auth.Authorization)
	}

	username, password, err := retrieveCredentials(authorization)
	if err != nil {
		return err
	}
	return b.users.check(ctx, username, password)
}

//retrieveCredentials 获取basicAuth认证信息
func retrieveCredentials(authInfo string) (string, string, error) {

	if authInfo != "" {
		const basic = "basic"
		l := len(basic)

		if len(authInfo) > l+1 && strings.ToLower(authInfo[:l]) == basic {
			b, err := base64.StdEncoding.DecodeString(authInfo[l+1:])
			if err != nil {
				return "", "", err
			}
			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					return cred[:i], cred[i+1:], nil
				}
			}
			return "", "", errors.New("[basic_auth] header has unrecognized format")
		}
		return "", "", errors.New("[basic_auth] header has unrecognized format")
	}
	return "", "", errors.New("[basic_auth] authorization required")
}
