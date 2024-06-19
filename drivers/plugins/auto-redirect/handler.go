package auto_redirect

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/upstream/balance"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*handler)(nil)
var _ eocontext.IFilter = (*handler)(nil)
var balanceFactory balance.IBalanceFactory
var discoveryAnonymous = static.CreateAnonymous(&static.Config{
	HealthOn: false,
	Health:   nil,
})

const maxRedirectCount = 10

//func init() {
//	balanceFactory, _ = balance.GetFactory("round-robin")
//}

type handler struct {
	drivers.WorkerBase
	maxRedirectCount int
	timeout          time.Duration
}

func (r *handler) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	if balanceFactory == nil {
		balanceFactory, _ = balance.GetFactory("round-robin")
	}
	return http_service.DoHttpFilter(r, ctx, next)
}

func (r *handler) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	if next != nil {
		err := next.DoChain(ctx)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	for i := 0; i < r.maxRedirectCount; i++ {
		if !fasthttp.StatusCodeIsRedirect(ctx.Response().StatusCode()) {
			return nil
		}
		err := redirect(ctx)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("too many redirects")
}

func redirect(ctx http_service.IHttpContext) error {
	location := ctx.Response().GetHeader("Location")
	u, err := url.Parse(location)
	if err != nil {
		return err
	}
	if u.Scheme != "" && u.Host != "" {
		app, err := discoveryAnonymous.GetApp(u.Host)
		if err != nil {
			return err
		}
		defer app.Close()

		balanceHandler, err := balanceFactory.Create(app, u.Scheme, ctx.GetBalance().TimeOut())
		if err != nil {
			return err
		}
		ctx.SetBalance(balanceHandler)
	}
	ctx.Proxy().URI().SetPath(u.RawPath)

	err = ctx.GetComplete().Complete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *handler) Destroy() {
}

func (r *handler) Start() error {
	return nil
}

func (r *handler) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}

	redirectCount := conf.MaxRedirectCount
	if redirectCount < 1 || redirectCount > maxRedirectCount {
		redirectCount = maxRedirectCount
	}
	r.maxRedirectCount = redirectCount

	return nil
}

func (r *handler) Stop() error {
	return nil
}

func (r *handler) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func getRedirectURL(baseURL string, location []byte) (string, string) {
	u := fasthttp.AcquireURI()
	u.Update(baseURL)
	u.UpdateBytes(location)
	u.RequestURI()
	defer fasthttp.ReleaseURI(u)
	return fmt.Sprintf("%s://%s", u.Scheme(), u.Host()), u.String()
}

func readPort(addr string) int {
	n := strings.LastIndex(addr, ":")
	if n >= 0 {
		p, e := strconv.Atoi(addr[n+1:])
		if e != nil {
			return p
		}
	}
	return 0
}

type redirectBalance struct {
}
