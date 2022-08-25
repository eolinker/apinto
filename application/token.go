package application

import (
	"errors"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/textproto"
)

var (
	PositionHeader = "header"
	PositionQuery  = "query"
)

var validPosition = []string{PositionHeader, PositionQuery}

func GetToken(ctx http_service.IHttpContext, tokenName string, position string) (string, bool) {
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

func HideToken(ctx http_service.IHttpContext, tokenName string, position string, ) {
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
