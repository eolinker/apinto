package request_file_parse

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"strings"

	"golang.org/x/text/encoding/charmap"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	MultipartForm = "multipart/form-data"
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
	if ctx.Request().Method() == "POST" || ctx.Request().Method() == "PUT" || ctx.Request().Method() == "PATCH" {
		contentType, _, err := mime.ParseMediaType(ctx.Request().ContentType())
		if err != nil {
			return err
		}
		if contentType == MultipartForm {
			// 当请求为文件请求时，解析文件
			fh, has := ctx.Request().Body().GetFile(e.fileKey)
			if has {
				for _, h := range fh {
					suffix, err := getFileSuffix(h)
					if err != nil {
						log.Errorf("get file suffix error: %v,name is %s", err, e.fileKey)
						continue
					}
					if _, ok := e.validSuf[suffix]; !ok {
						log.Errorf("file suffix is not valid,name is %s,suffix is %s", e.fileKey, suffix)
						continue
					}
					f, err := h.Open()
					if err != nil {
						log.Errorf("file open error: %v,name is %s", err, e.fileKey)
						continue
					}

					body, err := io.ReadAll(f)
					if err != nil {
						log.Errorf("read file body error: %v,name is %s", err, e.fileKey)
						f.Close()
						continue
					}
					f.Close()

					// body此处要做latin1编码
					out := make([]byte, 0, len(body))
					for _, t := range body {
						if v, ok := charmap.ISO8859_1.EncodeRune(rune(t)); ok {
							out = append(out, v)
						}
					}

					ctx.SetLabel("request_body", string(out))
					ctx.SetLabel("file_direction", "upload")
					ctx.SetLabel("file_name", h.Filename)
					ctx.SetLabel("file_suffix", suffix)
					ctx.WithValue("file_size", h.Size)

					break
				}
			}
		}
	}

	if next != nil {
		return next.DoChain(ctx)
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

func getFileSuffix(f *multipart.FileHeader) (string, error) {
	// 获取文件后缀
	fileName := f.Filename
	// 获取文件后缀
	suffix := fileName[strings.LastIndex(fileName, ".")+1:]
	if len(suffix) == 0 {
		contentType := f.Header.Get("Content-Type")
		if len(contentType) == 0 {
			return "", errors.New("file suffix is empty")
		}
	}
	return suffix, nil
}
