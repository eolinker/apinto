package jwt

import (
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/auth"
	http_context "github.com/eolinker/goku/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"jwt",
}

type jwt struct {
	id                string
	name              string
	credentials       *jwtUsers
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

	j.credentials = &jwtUsers{
		credentials: config.Credentials,
	}

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
	authorizationType, has := context.Request().Header().Get(auth.AuthorizationType)
	if !has {
		return auth.ErrorInvalidType
	}
	err := auth.CheckAuthorizationType(supportTypes, authorizationType)
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
