package jwt

import (
	"fmt"
	"github.com/eolinker/eosc/utils/config"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/eolinker/apinto/auth"
	"github.com/eolinker/eosc"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"jwt",
}

type jwt struct {
	id                string
	credentials       *jwtUsers
	signatureIsBase64 bool
	claimsToVerify    []string
	hideCredentials   bool
}

func (j *jwt) Id() string {
	return j.id
}

func (j *jwt) Start() error {
	return nil
}

func (j *jwt) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	c, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}

	j.credentials = &jwtUsers{
		credentials: c.Credentials,
	}

	j.signatureIsBase64 = c.SignatureIsBase64
	j.claimsToVerify = c.ClaimsToVerify
	j.hideCredentials = c.HideCredentials

	return nil
}

func (j *jwt) Stop() error {
	return nil
}

func (j *jwt) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (j *jwt) Auth(context http_service.IHttpContext) error {
	authorizationType := context.Request().Header().GetHeader(auth.AuthorizationType)
	if authorizationType == "" {
		return auth.ErrorInvalidType
	}
	err := auth.CheckAuthorizationType(supportTypes, authorizationType)
	if err != nil {
		return err
	}

	err = j.doJWTAuthentication(context)
	if err != nil {
		return err
	}

	return nil
}
