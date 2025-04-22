package aes

import (
	"net/http"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eosc.IWorker = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	encodeFunc func(string) (string, error)
	decodeFunc func(string) (string, error)
}

func (e *executor) Destroy() {
	e.encodeFunc = nil
	e.decodeFunc = nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if ctx.Request().Method() == http.MethodPost || ctx.Request().Method() == http.MethodPut || ctx.Request().Method() == http.MethodPatch {
		body, _ := ctx.Proxy().Body().RawBody()
		decodeBody, err := e.decodeFunc(string(body))
		if err != nil {
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetHeader("Content-Type", "text/plain")
			ctx.Response().SetBody([]byte("decode body error"))
			return err
		}
		ctx.SetLabel("request_body", decodeBody)
		ctx.WithValue("request_body_complete", 1)
		ctx.Proxy().Body().SetRaw("application/json", []byte(decodeBody))
	}

	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	body := ctx.Response().GetBody()

	encodeBody, err := e.encodeFunc(string(body))
	if err != nil {
		return err
	}
	ctx.Response().SetBody([]byte(encodeBody))
	ctx.Response().SetHeader("Content-Type", "text/plain")
	ctx.Response().SetStatus(200, "200")
	return nil
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset(conf *Config) error {
	keyLength := 16
	switch conf.Algorithm {
	case "AES-128":
		keyLength = 16
	case "AES-192":
		keyLength = 24
	case "AES-256":
		keyLength = 32
	}
	cipher, err := NewAESCipher(conf.Mode, conf.Key, keyLength)
	if err != nil {
		return err
	}
	switch conf.Mode {
	case "ECB":
		e.encodeFunc = cipher.EncryptECB
		e.decodeFunc = cipher.DecryptECB
	case "CBC":
		e.encodeFunc = cipher.EncryptCBC
		e.decodeFunc = cipher.DecryptCBC
	case "CTR":
		e.encodeFunc = cipher.EncryptCTR
		e.decodeFunc = cipher.DecryptCTR
	case "OFB":
		e.encodeFunc = cipher.EncryptOFB
		e.decodeFunc = cipher.DecryptOFB
	case "CFB":
		e.encodeFunc = cipher.EncryptCFB
		e.decodeFunc = cipher.DecryptCFB
	}

	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
