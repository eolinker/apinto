package session_keep

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

const SessionName = "Apinto-Session"

type balanceSelectKeyType struct {
}

var (
	balanceFirstSelectKey = balanceSelectKeyType{}
)

type Session struct {
	eocontext.BalanceHandler
}

func NewSession(base eocontext.BalanceHandler) eocontext.BalanceHandler {
	return &Session{BalanceHandler: base}
}

func (s *Session) Select(ctx eocontext.EoContext) (eocontext.INode, int, error) {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return s.BalanceHandler.Select(ctx)
	}
	value := httpContext.Value(balanceFirstSelectKey)
	if value != nil {
		// 请求重试，重新选节点
		return s.BalanceHandler.Select(ctx)
	}
	httpContext.WithValue(balanceFirstSelectKey, true)

	session := httpContext.Request().Header().GetCookie(SessionName)
	if session != "" {
		index := httpContext.Request().Header().GetCookie(fmt.Sprintf("Apinto-Upstream-%s", session))
		if index != "" {
			indexV, _ := strconv.Atoi(index)
			app := httpContext.GetBalance()
			nodes := app.Nodes()
			if indexV < len(nodes) && nodes[indexV].Status() == eocontext.Running {
				return nodes[indexV], indexV, nil
			}
		}
	}

	node, i, err := s.BalanceHandler.Select(httpContext)
	if err != nil {
		return nil, 0, err
	}

	if session == "" {
		session = ctx.RequestId()
		cookieSession := http.Cookie{Name: SessionName, Value: session}
		httpContext.Response().AddHeader("Set-Cookie", cookieSession.String())
	}
	indexCookie := http.Cookie{Name: fmt.Sprintf("Apinto-Upstream-%s", session), Value: strconv.Itoa(i)}

	httpContext.Response().AddHeader("Set-Cookie", indexCookie.String())
	return node, i, nil
}
