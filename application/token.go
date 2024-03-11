package application

import (
	"errors"
	"net/textproto"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	PositionHeader = "header"
	PositionQuery  = "query"
	PositionBody   = "body"
)

var validPosition = []string{PositionHeader, PositionQuery, PositionBody}

//func GetToken(ctx http_service.IHttpContext, tokenName string, position string) (string, bool) {
//	token, has := getToken(ctx, tokenName, position)
//	if has {
//		ctx.SetLabel("token", token)
//	}
//	return token, has
//}

func GetToken(ctx http_service.IHttpContext, tokenName string, position string) (string, bool) {
	switch position {
	case PositionHeader:
		value, has := ctx.Request().Header().Headers()[textproto.CanonicalMIMEHeaderKey(tokenName)]
		if has {
			return value[0], has
		}
		return "", false
	case PositionQuery:
		value := ctx.Request().URI().GetQuery(tokenName)
		return value, true
	case "":
		{
			value, has := ctx.Request().Header().Headers()["Authorization"]
			if has {
				return value[0], has
			}
			return "", false
		}
	}
	return "", false
}

func HideToken(ctx http_service.IHttpContext, tokenName string, position string) {
	switch position {
	case PositionHeader:
		ctx.Proxy().Header().DelHeader(textproto.CanonicalMIMEHeaderKey(tokenName))
	case PositionQuery:
		ctx.Proxy().URI().DelQuery(tokenName)
	case "":
		{
			ctx.Proxy().Header().DelHeader("Authorization")
		}
	}
	return
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
