package rewrite

import (
	"fmt"
	"strings"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/service"
)

type Rewrite struct {
	*Driver
	id   string
	name string
	path string
}

func (r *Rewrite) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	router := getEndpoint(ctx)
	if router != nil {
		// 设置目标URL
		location, has := router.Location()

		if has && location.CheckType() == http_service.CheckTypePrefix {
			path := recombinePath(string(ctx.Request().URL().Path), location.Value(), r.path)
			ctx.Request().URL().Path = path
		}
	} else {
		if r.path != "" {
			ctx.Request().URL().Path = r.path
		}
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (r *Rewrite) Id() string {
	return r.id
}

func (r *Rewrite) Start() error {
	return nil
}

func (r *Rewrite) Reset(v interface{}, workers map[eosc.RequireId]interface{}) error {
	conf, err := r.check(v)
	if err != nil {
		return err
	}
	r.path = conf.ReWriteUrl
	return nil
}

func (r *Rewrite) Stop() error {
	return nil
}

func (r *Rewrite) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func getEndpoint(ctx http_service.IHttpContext) service.IRouterEndpoint {
	value := ctx.Value("router.endpoint")
	if value == nil {
		return nil
	}
	if router, ok := value.(service.IRouterEndpoint); ok {
		return router
	}
	return nil
}

//recombinePath 生成新的目标URL
func recombinePath(requestURL, location, targetURL string) string {
	newRequestURL := strings.Replace(requestURL, location, "", 1)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(targetURL, "/"), strings.TrimPrefix(newRequestURL, "/"))
}
