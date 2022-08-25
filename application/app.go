package application

import (
	"errors"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/utils/config"
)

var (
	appSkill         string
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token is expired")
)

func init() {
	var t IApp
	appSkill = config.TypeNameOf(&t)
}

type IApp interface {
	IAuthUser
}

func CheckSkill(skill string) bool {
	return skill == appSkill
}

type IAuth interface {
	ID() string
	Check(appID string, users []*User) error
	Set(appID string, labels map[string]string, disable bool, users []*User)
	Del(appID string)
	UserCount() int
	IAuthUser
}

type IAuthUser interface {
	Driver() string
	GetUser(ctx http_service.IHttpContext) (*UserInfo, bool)
}
