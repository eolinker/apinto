package servicehttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/upstream"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/service"
)

type serviceWorker struct {
	id         string
	name       string
	driver     string
	desc       string
	timeout    time.Duration
	rewriteUrl string
	retry      int
	scheme     string
	proxyAddr  string
	upstream   upstream.IUpstream
}

func (s *serviceWorker) Id() string {
	return s.id
}

func (s *serviceWorker) Start() error {
	return nil
}

func (s *serviceWorker) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	data, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), eosc.ErrorStructType)
	}
	if worker, has := workers[data.Upstream]; has {
		s.desc = data.Desc
		s.timeout = time.Duration(data.Timeout) * time.Millisecond
		s.rewriteUrl = data.RewriteURL
		s.retry = data.Retry
		s.scheme = data.Scheme
		u, ok := worker.(upstream.IUpstream)
		if ok {
			s.upstream = u
			return nil
		}
	} else {
		worker, has = workers[eosc.RequireId(fmt.Sprintf("%s@%s", data.Upstream, "upstream"))]
		if has {
			s.desc = data.Desc
			s.timeout = time.Duration(data.Timeout) * time.Millisecond
			s.rewriteUrl = data.RewriteURL
			s.retry = data.Retry
			s.scheme = data.Scheme
			u, ok := worker.(upstream.IUpstream)
			if ok {
				s.upstream = u
				return nil
			}
			return nil
		}
	}
	return errors.New("fail to create serviceWorker")

}

func (s *serviceWorker) Stop() error {
	return nil
}

func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}

func (s *serviceWorker) Name() string {
	return s.name
}

func (s *serviceWorker) Desc() string {
	return s.desc
}

func (s *serviceWorker) Retry() int {
	return s.retry
}

func (s *serviceWorker) Timeout() time.Duration {
	return s.timeout
}

func (s *serviceWorker) Scheme() string {
	return s.scheme
}

func (s *serviceWorker) ProxyAddr() string {
	return s.proxyAddr
}

func (s *serviceWorker) Handle(w http.ResponseWriter, r *http.Request, router service.IRouterRule) error {
	// 构造context
	ctx := http_context.NewContext(r, w)
	// 设置目标URL
	ctx.ProxyRequest.SetTargetURL(recombinePath(r.URL.Path, router.Location(), s.rewriteUrl))
	s.upstream.Send(ctx, s)
	return nil
}

func recombinePath(requestURL, location, targetURL string) string {
	new := strings.Replace(requestURL, location, "", 1)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(targetURL, "/"), strings.TrimPrefix(new, "/"))
}
