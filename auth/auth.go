package auth

import (
	"errors"

	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

const (
	//AuthorizationType 鉴权类型
	AuthorizationType = "Authorization-Type"
	//Authorization 鉴权
	Authorization = "Authorization"
)

//ErrorInvalidType 非法的鉴权类型
var ErrorInvalidType = errors.New("invalid authorization type")

//CheckSkill 检查能力
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku-eosc/auth.auth.IAuth"
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
	for _, t := range supportTypes {
		if t == authType {
			return nil
		}
	}
	return ErrorInvalidType
}
