package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/application/auth/oauth2"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eocontext.IFilter = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	cfg   *Config
	cache scope_manager.IProxyOutput[resources.ICache]
	once  sync.Once
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if !strings.HasSuffix(ctx.Request().URI().Path(), "/oauth2/token") && !strings.HasSuffix(ctx.Request().URI().Path(), "/oauth2/authorize") {
		if next != nil {
			err = next.DoChain(ctx)
		}
		return
	}

	params := retrieveParameters(ctx)
	clientId := params.Get("client_id")
	if clientId == "" {
		// 当空时视为正常请求，不做拦截
		if next != nil {
			err = next.DoChain(ctx)
		}
		return
	}
	defer func() {
		if err != nil {
			log.Error(err)
			type errResp struct {
				Message string `json:"message"`
			}
			msg, _ := json.Marshal(errResp{Message: "Unauthorized"})
			ctx.Response().SetBody(msg)
			ctx.Response().SetStatus(http.StatusUnauthorized, "unauthorized")
		}
	}()
	client, has := oauth2.GetClient(clientId)
	if !has {
		err = fmt.Errorf("invalid client id")
		return
	}

	if strings.ToUpper(ctx.Request().URI().Scheme()) != "HTTPS" && !e.cfg.AcceptHttpIfAlreadyTerminated {
		err = fmt.Errorf("invalid scheme")
		return
	}
	if client.Expire() > 0 && client.Expire() < time.Now().Unix() {
		err = fmt.Errorf("client id is expired")
		return
	}

	e.once.Do(func() {
		e.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})

	var data []byte
	if strings.HasSuffix(ctx.Request().URI().Path(), "/oauth2/authorize") {
		data, err = e.Authorize(ctx, client, params)
	} else if strings.HasSuffix(ctx.Request().URI().Path(), "/oauth2/token") {
		data, err = e.Token(ctx, client, params)
	}
	if err != nil {
		return
	}
	ctx.Response().SetBody(data)
	ctx.Response().SetStatus(http.StatusOK, "ok")
	return
}

func (e *executor) Destroy() {
	return
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
