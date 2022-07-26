package proxy_rewrite2

import (
	"fmt"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"regexp"
)

var _ http_service.IFilter = (*ProxyRewrite)(nil)

type ProxyRewrite struct {
	*Driver
	id          string
	pathType    string
	staticPath  string
	prefixPath  []*SPrefixPath
	regexPath   []*SRegexPath
	regexMatch  []*regexp.Regexp
	hostRewrite bool
	host        string
	headers     map[string]string
}

func (p *ProxyRewrite) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	err = p.rewrite(ctx)
	if err != nil {
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (p *ProxyRewrite) rewrite(ctx http_service.IHttpContext) error {
	//修改scheme
	if p.scheme != "" {
		ctx.Proxy().URI().SetScheme(p.scheme)
	}

	//修改uri   uri比regexURI更优先
	if p.uri != "" {
		ctx.Proxy().URI().SetPath(p.uri)
	} else if p.regexMatch != nil {
		requestURI := ctx.Proxy().URI().Path()
		if p.regexMatch.MatchString(requestURI) {
			newURI := p.regexMatch.ReplaceAllString(requestURI, p.regexURI[1])
			ctx.Proxy().URI().SetPath(newURI)
		}
	}

	//修改header中的host
	if p.host != "" {
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

	return nil
}

func (p *ProxyRewrite) Id() string {
	return p.id
}

func (p *ProxyRewrite) Start() error {
	return nil
}

func (p *ProxyRewrite) Reset(v interface{}, workers map[eosc.RequireId]interface{}) error {
	conf, err := p.check(v)
	if err != nil {
		return err
	}

	p.pathType = conf.PathType
	p.hostRewrite = conf.HostRewrite
	p.host = conf.Host
	p.headers = conf.Headers

	switch conf.PathType {
	case "static":
		p.staticPath = conf.StaticPath
	case "prefix":
		p.prefixPath = conf.PrefixPath
	case "regex":
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
	p.pathType = "none"
	p.hostRewrite = false
	p.host = ""
	p.staticPath = ""
	p.prefixPath = nil
	p.regexPath = nil
	p.regexMatch = nil
	p.headers = nil
}

func (p *ProxyRewrite) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
