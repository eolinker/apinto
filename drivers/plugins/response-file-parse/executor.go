package response_file_parse

import (
	"strings"

	"golang.org/x/text/encoding/charmap"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	//csv,tar,bz2,xz,jar,pdf,doc,docx,xls,ppt,xlsx,pptx,zip,txt,rar,gz,dot
	defaultValidSuf = map[string]struct{}{
		"csv":  {},
		"tar":  {},
		"bz2":  {},
		"xz":   {},
		"jar":  {},
		"pdf":  {},
		"doc":  {},
		"docx": {},
		"xls":  {},
		"ppt":  {},
		"xlsx": {},
		"pptx": {},
		"zip":  {},
		"txt":  {},
		"rar":  {},
		"gz":   {},
		"dot":  {},
	}
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	fileKey      string
	validSuf     map[string]struct{}
	largeWarn    int64
	largeWarnStr string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	contentDisposition := ctx.Response().Headers().Get("Content-Disposition")
	if contentDisposition != "" {
		params := strings.Split(contentDisposition, ";")
		paramsMap := make(map[string]string, len(params))
		for _, param := range params {
			param = strings.TrimSpace(param)
			ps := strings.Split(param, "=")
			if ps[0] != "" {
				if len(ps) > 1 {
					paramsMap[ps[0]] = ps[1]
				} else {
					paramsMap[ps[0]] = ""
				}
			}
		}
		if err != nil {
			log.Errorf("parse content disposition error: %v", err)
			return nil
		}
		if fileName, ok := paramsMap[e.fileKey]; ok {
			if fileName != "" {
				suffix := fileName[strings.LastIndex(fileName, ".")+1:]
				if _, ok := e.validSuf[suffix]; !ok {
					log.Errorf("file suffix is not valid,name is %s,suffix is %s", e.fileKey, suffix)
					return nil
				}
				body := ctx.Response().GetBody()
				// body此处要做latin1编码
				out := make([]byte, 0, len(body))
				for _, t := range body {
					if v, ok := charmap.ISO8859_1.EncodeRune(rune(t)); ok {
						out = append(out, v)
					}
				}
				size := len(out)
				ctx.WithValue("response_body", string(out))
				ctx.WithValue("file_direction", "download")
				ctx.WithValue("file_name", fileName)
				ctx.WithValue("file_suffix", suffix)
				ctx.WithValue("file_size", size)
				if int64(size) > e.largeWarn {
					ctx.WithValue("file_large_warn", e.largeWarnStr)
				}
			}
		}
	}
	return nil
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
