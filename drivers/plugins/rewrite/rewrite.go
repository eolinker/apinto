package rewrite

import (
	"fmt"
	"strings"

	"github.com/eolinker/goku/checker"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/service"
)

var _ http_service.IFilter = (*Rewrite)(nil)

type Rewrite struct {
	*Driver
	id   string
	name string
	path string
}

func (r *Rewrite) Destroy() {

}

func (r *Rewrite) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	router, has := service.EndpointFromContext(ctx)
	if has {

		if router != nil {
			// 设置目标URL
			location, has := router.Location()

			if has && location.CheckType() == checker.CheckTypePrefix {
				ctx.Proxy().SetPath(recombinePath(string(ctx.Request().URL().Path), location.Value(), r.path))
			}
		} else {
			if r.path != "" {
				ctx.Proxy().SetPath(r.path)
			}
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

//recombinePath 生成新的目标URL
func recombinePath(requestURL, location, targetURL string) string {
	newRequestURL := strings.Replace(requestURL, location, "", 1)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(targetURL, "/"), strings.TrimPrefix(newRequestURL, "/"))
}
