package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/eolinker/eosc"

	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"basic",
	"basic_auth",
	"basic-auth",
	"basicauth",
}

type basic struct {
	id     string
	name   string
	driver string
	users  []User
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
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}
	b.users = cfg.User
	return nil
}

func (b *basic) Stop() error {
	return nil
}

func (b *basic) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (b *basic) Auth(ctx *http_context.Context) error {
	err := auth.CheckAuthorizationType(supportTypes, ctx.Request().Headers().Get(auth.AuthorizationType))
	if err != nil {
		return err
	}
	authorization := ctx.Request().Headers().Get(auth.Authorization)
	username, password, err := retrieveCredentials(authorization)
	if err != nil {
		return err
	}
	for _, u := range b.users {
		if u.Username == username && u.Password == password {
			return nil
		}
	}
	return errors.New("invalid user")
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
