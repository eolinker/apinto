package proxy_rewrite_v2

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"regexp"
	"strings"
)

var _ eocontext.IFilter = (*ProxyRewrite)(nil)
var _ http_service.HttpFilter = (*ProxyRewrite)(nil)

const (
	typeNone   = "none"
	typeStatic = "static"
	typePrefix = "prefix"
	typeRegex  = "regex"
)

var (
	regexpErrInfo   = `[plugin proxy-rewrite2 config err] Compile regexp fail. err regexp: %s `
	hostErrInfo     = `[plugin proxy-rewrite2 config err] Host can't be null. `
	notMatchErrInfo = `[plugin proxy-rewrite2 err] Proxy path rewrite fail. Request path can't match any rewrite-path. request path: %s `
)

type ProxyRewrite struct {
	*Driver
	id          string
	pathType    string
	staticPath  string
	prefixPath  []*SPrefixPath
	regexPath   []*SRegexPath
	regexMatch  []*regexp.Regexp
	notMatchErr bool
	hostRewrite bool
	host        string
	headers     map[string]string
}

func (p *ProxyRewrite) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(p, ctx, next)
}

func (p *ProxyRewrite) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	isPathMatch := p.rewrite(ctx)
	if p.notMatchErr && !isPathMatch {
		err := fmt.Errorf(notMatchErrInfo, ctx.Proxy().URI().Path())
		ctx.Response().SetStatus(400, "400")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (p *ProxyRewrite) rewrite(ctx http_service.IHttpContext) bool {
	//修改header中的host
	if p.hostRewrite {
		ctx.Proxy().URI().SetHost(p.host)
	}

	//修改转发至上游的header，v可设置为空字符串，此时代表删掉header中对应的key. 若header某个key已存在则重写
	for k, v := range p.headers {
		if v == "" {
			ctx.Proxy().Header().DelHeader(k)
			continue
		}
		ctx.Proxy().Header().SetHeader(k, v)
	}

	pathMatch := false
	switch p.pathType {
	case typeStatic:
		ctx.Proxy().URI().SetPath(p.staticPath)
		pathMatch = true
	case typePrefix:
		oldPath := ctx.Proxy().URI().Path()
		for _, pPath := range p.prefixPath {
			if strings.HasPrefix(oldPath, pPath.PrefixPathMatch) {
				newPath := strings.Replace(oldPath, pPath.PrefixPathMatch, pPath.PrefixPathReplace, 1)
				ctx.Proxy().URI().SetPath(newPath)
				pathMatch = true
				break
			}
		}

	case typeRegex:
		oldPath := ctx.Proxy().URI().Path()
		for i, rPath := range p.regexPath {
			reg := p.regexMatch[i]
			if reg.MatchString(oldPath) {
				newPath := reg.ReplaceAllString(oldPath, rPath.RegexPathReplace)
				ctx.Proxy().URI().SetPath(newPath)
				pathMatch = true
				break
			}
		}

	case typeNone:
		pathMatch = true
	}

	return pathMatch
}

func (p *ProxyRewrite) Id() string {
	return p.id
}

func (p *ProxyRewrite) Start() error {
	return nil
}

func (p *ProxyRewrite) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := p.check(v)
	if err != nil {
		return err
	}

	p.pathType = conf.PathType
	p.notMatchErr = conf.NotMatchErr
	p.hostRewrite = conf.HostRewrite
	p.host = conf.Host
	p.headers = conf.Headers

	switch conf.PathType {
	case typeStatic:
		p.staticPath = conf.StaticPath
	case typePrefix:
		p.prefixPath = conf.PrefixPath
	case typeRegex:
		regexMatch := make([]*regexp.Regexp, 0)

		for _, rPath := range conf.RegexPath {
			rMatch, err := regexp.Compile(rPath.RegexPathMatch)
			if err != nil {
				return fmt.Errorf(regexpErrInfo, rPath.RegexPathMatch)
			}
			regexMatch = append(regexMatch, rMatch)
		}
		p.regexPath = conf.RegexPath
		p.regexMatch = regexMatch
	}

	return nil
}

func (p *ProxyRewrite) Stop() error {
	return nil
}

func (p *ProxyRewrite) Destroy() {
	p.pathType = typeNone
	p.hostRewrite = false
	p.host = ""
	p.staticPath = ""
	p.prefixPath = nil
	p.regexPath = nil
	p.regexMatch = nil
	p.notMatchErr = false
	p.headers = nil
}

func (p *ProxyRewrite) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
