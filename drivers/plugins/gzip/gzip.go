package gzip

import (
	"bytes"
	"compress/gzip"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"strings"
)

var _ http_service.HttpFilter = (*Gzip)(nil)
var _ eocontext.IFilter = (*Gzip)(nil)

type Gzip struct {
	drivers.WorkerBase
	conf *Config
}

func (g *Gzip) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(g, ctx, next)
}

func (g *Gzip) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	head := ctx.Request().Header().GetHeader("Accept-Encoding")
	if next != nil {
		err = next.DoChain(ctx)
	}
	if err == nil && strings.Contains(head, "gzip") {
		err = g.doCompress(ctx)
	}
	return
}

func (g *Gzip) doCompress(ctx http_service.IHttpContext) error {
	flag := false
	resp := ctx.Response()

	if resp.BodyLen() < g.conf.MinLength {
		// 小于要求的最低长度，不压缩
		return nil
	}
	contentType := resp.GetHeader("Content-Type")
	if len(g.conf.Types) == 0 {
		flag = true
	} else {
		for _, t := range g.conf.Types {
			if strings.Contains(contentType, t) {
				flag = true
				break
			}
		}
	}
	if flag {
		res, err := g.compress(resp.GetBody())
		if err != nil {
			return err
		}
		resp.SetBody(res)
		resp.SetHeader("Content-Encoding", "gzip")
		if g.conf.Vary {
			resp.SetHeader("Vary", "Accept-Encoding")
		}
	}
	return nil
}
func (g *Gzip) compress(content []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(content)
	if err != nil {
		return nil, err
	}
	err = zw.Flush()
	if err != nil {
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *Gzip) Destroy() {
	g.conf = nil
}

func (g *Gzip) Start() error {
	return nil
}

func (g *Gzip) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	g.conf = cfg
	return nil
}

func (g *Gzip) Stop() error {
	return nil
}

func (g *Gzip) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
