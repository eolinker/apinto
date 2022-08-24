package application

import (
	"errors"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	appSkill         string
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token is expired")
)

//func init() {
//	var t IApp
//	appSkill = config.TypeNameOf(&t)
//}
//
//type IApp interface {
//	Auth(ctx eocontext.EoContext) error
//}
//
//func CheckSkill(skill string) bool {
//	return skill == appSkill
//}

type IAuth interface {
	ID() string
	Driver() string
	Check(appID string, users []*User) error
	Set(appID string, labels map[string]string, disable bool, users []*User)
	Del(appID string)
	UserCount() int
	IAuthFilter
}

type IAuthFilter interface {
	Auth(ctx http_service.IHttpContext) error
}
