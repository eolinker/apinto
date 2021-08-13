package auth

import (
	"errors"
	"strings"

	http_context "github.com/eolinker/goku/node/http-context"
)

const (
	//AuthorizationType 鉴权类型
	AuthorizationType = "Authorization-Type"
	//Authorization 鉴权
	Authorization = "Authorization"
)

var (
	//ErrorInvalidType 非法的鉴权类型
	ErrorInvalidType = errors.New("invalid authorization type")

	//ErrorInvalidUser 非法用户
	ErrorInvalidUser = errors.New("invalid user")

	//ErrorExpireUser 用户已过期
	ErrorExpireUser = errors.New("the user is expired")
)

//CheckSkill 检查能力
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku/auth.auth.IAuth"
}

//IAuth 鉴权接口声明
type IAuth interface {
	Auth(ctx *http_context.Context) error
}

//CheckAuthorizationType 检查鉴权类型是否合法
func CheckAuthorizationType(supportTypes []string, authType string) error {
	if authType == "" {
		return ErrorInvalidType
	}
	authType = strings.ToLower(authType)
	for _, t := range supportTypes {
		if t == authType {
			return nil
		}
	}
	return ErrorInvalidType
}
