package auth

import (
	"errors"

	"github.com/eolinker/apinto/auth"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
)

var _ http_service.IFilter = (*Auth)(nil)

type Auth struct {
	*Driver
	id    string
	auths []auth.IAuth
}

func (a *Auth) Destroy() {

}

func (a *Auth) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) error {
	err := a.doAuth(ctx)
	if err != nil {
		resp := ctx.Response()

		resp.SetBody([]byte(err.Error()))
		resp.SetStatus(403, "403")
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (a *Auth) doAuth(ctx http_service.IHttpContext) error {
	// 鉴权
	auths := a.auths
	if len(auths) > 0 {
		validRequest := false
		for _, a := range auths {
			err := a.Auth(ctx)
			if err == nil {
				validRequest = true
				break
			}
			log.Error(err)
		}
		if !validRequest {
			return errors.New("invalid user")
		}
	}
	return nil
}
func (a *Auth) Id() string {
	return a.id
}

func (a *Auth) Start() error {
	return nil
}

func (a *Auth) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	confObj, err := a.check(conf)
	if err != nil {
		return err
	}
	list, err := a.getList(confObj.Auth)
	if err != nil {
		return err
	}

	a.auths = list
	return nil
}

func (a *Auth) Stop() error {
	return nil
}

func (a *Auth) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
