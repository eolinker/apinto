package jwt

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"Jwt",
	"jwt",
	"JWT",
}

type jwt struct {
	id                string
	name              string
	credentials       map[string]*JwtCredential //key为iss和algorithm拼接的字符串，且唯一
	signatureIsBase64 bool
	claimsToVerify    []string
	runOnPreflight    bool
	hideCredentials   bool
}

func (j *jwt) Id() string {
	return j.id
}

func (j *jwt) Start() error {
	return nil
}

func (j *jwt) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}

	//验证证书列表里每个证书的iss和algorithm组合是唯一的,并且返回Key为IssAlgorithm,Value为Credential的map
	credentials, err := validateCredentials(config.Credentials)
	if err != nil {
		return err
	}

	j.credentials = credentials
	j.signatureIsBase64 = config.SignatureIsBase64
	j.claimsToVerify = config.ClaimsToVerify
	j.runOnPreflight = config.RunOnPreflight
	j.hideCredentials = config.HideCredentials

	return nil
}

func (j *jwt) Stop() error {
	return nil
}

func (j *jwt) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (j *jwt) Auth(context *http_context.Context) error {
	err := auth.CheckAuthorizationType(supportTypes, context.Request().Headers().Get(auth.AuthorizationType))
	if err != nil {
		return err
	}

	if !j.runOnPreflight && context.Request().Method() == "OPTIONS" {
		return nil
	}
	err = j.doJWTAuthentication(context)
	if err != nil {
		return err
	}

	return nil
}
