package grey_strategy

import (
	"fmt"
	"net/http"

	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

const SessionName = "Apinto-Session"

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

	session := httpCtx.Request().Header().GetCookie(SessionName)
	var cookieKey string
	if len(session) != 0 {
		cookieKey = fmt.Sprintf(cookieName, session)
		cookie := httpCtx.Request().Header().GetCookie(cookieKey)
		if cookie == grey {
			return true
		} else if cookie == normal {
			return false
		}
	}
	if session == "" {
		session = ctx.RequestId()
		cookieSession := http.Cookie{Name: SessionName, Value: session}
		cookieKey = fmt.Sprintf(cookieName, session)
		httpCtx.Response().AddHeader("Set-Cookie", cookieSession.String())
	}

	ok := k.GreyMatch.Match(ctx)
	cookie := normal
	if ok {
		cookie = grey
	}
	httpCtx.Response().AddHeader("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, cookie))
	return ok

}
