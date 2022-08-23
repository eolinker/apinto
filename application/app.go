package application

import (
	"errors"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/textproto"
)

var (
	appSkill         string
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")
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

type User struct {
	Expire         int64             `json:"expire"`
	Labels         map[string]string `json:"labels"`
	Pattern        map[string]string `json:"pattern"`
	HideCredential bool              `json:"hide_credential"`
}

type IAuth interface {
	ID() string
	Driver() string
	Set(appID string, users []*User)
	Del(appID string)
	IAuthFilter
}

type IAuthFilter interface {
	Auth(ctx http_service.IHttpContext) error
}

var (
	PositionHeader = "header"
	PositionQuery  = "query"
)

var validPosition = []string{PositionHeader, PositionQuery}

func GetToken(tokenName string, position string, ctx http_service.IHttpContext) (string, bool) {
	switch position {
	case PositionHeader:
		value, has := ctx.Request().Header().Headers()[textproto.CanonicalMIMEHeaderKey(tokenName)]
		return value[0], has
	case PositionQuery:
		value := ctx.Request().URI().GetQuery(tokenName)
		return value, true
	case "":
		{
			value, has := ctx.Request().Header().Headers()["Authorization"]
			return value[0], has
		}
	}
	return "", false
}

func CheckPosition(position string) error {
	if position == "" {
		return nil
	}
	for _, p := range validPosition {
		if p == position {
			return nil
		}
	}
	return errors.New("invalid position: " + position)
}
