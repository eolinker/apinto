package acl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
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
	allow map[string]struct{}
	deny  map[string]struct{}
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	groupMap := make(map[string]struct{})
	groups := ctx.Value("acl_groups")
	groupArr := make([]string, 0)
	switch gs := groups.(type) {
	case []interface{}:
		for _, g := range gs {
			switch t := g.(type) {
			case string:
				groupMap[t] = struct{}{}
				groupArr = append(groupArr, t)
			case []interface{}:
				for _, v := range t {
					if s, ok := v.(string); ok {
						groupMap[s] = struct{}{}
						groupArr = append(groupArr, s)
					}
				}
			}
		}
	case []string:
		for _, g := range gs {
			groupMap[g] = struct{}{}
			groupArr = append(groupArr, g)
		}
	case string:
		groupMap[gs] = struct{}{}
		groupArr = append(groupArr, gs)
	}
	type Msg struct {
		Message string `json:"message"`
	}
	for key := range e.deny {
		if _, ok := groupMap[key]; ok {
			// 拒绝访问
			err = fmt.Errorf("acl deny, group is %s", key)
			log.Error(err)
			msg := Msg{
				Message: "Unauthorized",
			}
			data, _ := json.Marshal(msg)
			ctx.Response().SetStatus(401, "Unauthorized")
			ctx.Response().SetBody([]byte(data))
			return
		}
	}
	allow := false
	for key := range e.allow {
		if _, ok := groupMap[key]; ok {
			allow = true
			break
		}
	}
	if !allow {
		err = fmt.Errorf("groups is not allow, groups is %v", groups)
		log.Error(err)
		msg := Msg{
			Message: "Unauthorized",
		}
		data, _ := json.Marshal(msg)
		ctx.Response().SetStatus(401, "Unauthorized")
		ctx.Response().SetBody([]byte(data))
		return
	}

	if !e.cfg.HideGroupsHeader {
		ctx.Proxy().Header().SetHeader("X-Consumer-Groups", strings.Join(groupArr, ","))
	}
	if next != nil {
		return next.DoChain(ctx)
	}
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
