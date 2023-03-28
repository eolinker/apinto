package grey_strategy

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type keepSessionGreyFlow struct {
	GreyMatch
}

// Match 保持会话连接
func (k *keepSessionGreyFlow) Match(ctx eocontext.EoContext) bool {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		log.Error("keepSessionGreyFlow err=%s", err.Error())
		return false
	}

	session := httpCtx.Request().Header().GetCookie("session")
	if len(session) == 0 {
		return k.GreyMatch.Match(ctx)
	}

	cookieKey := fmt.Sprintf(cookieName, session)

	cookie := httpCtx.Request().Header().GetCookie(cookieKey)
	if cookie == grey {
		return true
	} else if cookie == normal {
		return false
	}

	if k.GreyMatch.Match(ctx) {
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, grey))
		return true
	} else {
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, normal))
		return false
	}
}
