package rsa_filter

import (
	"crypto/rsa"
	"fmt"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eocontext.IFilter = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	// 解密 + 签名
	privateKey *rsa.PrivateKey
	// 验签 + 加密
	publicKey          *rsa.PublicKey
	requestSignHeader  string
	responseSignHeader string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	orgContentType := ctx.Request().Header().GetHeader("Origin-Content-Type")
	if orgContentType == "" {
		orgContentType = "application/json"
	}

	body, _ := ctx.Request().Body().RawBody()
	signature := ctx.Request().Header().GetHeader(e.requestSignHeader)
	// 解密
	decrBody, err := decrypt(body, e.privateKey)
	if err != nil {
		err = fmt.Errorf("decrypt body error: %w", err)
		ctx.Response().SetBody([]byte(err.Error()))
		ctx.Response().SetStatus(403, "403")
		return err
	}
	ctx.SetLabel("request_body", string(decrBody))
	ctx.WithValue("request_body_complete", 1)
	ctx.Request().Header().Headers().Set("Content-Type", orgContentType)
	if signature != "" {
		decodeSign, err := utils.B64Decode(signature)
		if err != nil {
			err = fmt.Errorf("base64 decode error: %w", err)
			ctx.Response().SetBody([]byte(err.Error()))
			ctx.Response().SetStatus(403, "403")
			return err
		}
		// 验签
		err = verify(body, decodeSign, e.publicKey)
		if err != nil {
			err = fmt.Errorf("verify signature error: %w", err)
			ctx.Response().SetBody([]byte(err.Error()))
			ctx.Response().SetStatus(403, "403")
			return err
		}
	}

	// 转发时传输明文
	ctx.Proxy().Body().SetRaw(orgContentType, decrBody)
	ctx.Proxy().Header().SetHeader("Content-Type", orgContentType)
	if next != nil {
		err = next.DoChain(ctx)
	}

	responseContentType := ctx.Response().GetHeader("Content-Type")
	body = ctx.Response().GetBody()
	// 加密
	encBody, err := encrypt(body, e.publicKey)
	if err != nil {
		err = fmt.Errorf("encrypt body error: %w", err)
		ctx.Response().SetBody([]byte(err.Error()))
		ctx.Response().SetStatus(403, "403")
		return
	}
	if e.responseSignHeader != "" {
		responseSign, err := sign(encBody, e.privateKey)
		if err != nil {
			err = fmt.Errorf("sign body error: %w", err)
			ctx.Response().SetBody([]byte(err.Error()))
			ctx.Response().SetStatus(403, "403")
			return err
		}
		ctx.Response().SetHeader(e.responseSignHeader, utils.B64Encode(responseSign))
	}

	ctx.Response().SetBody(encBody)
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Origin-Content-Type", responseContentType)
	return
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

func (e *executor) reset(conf *Config) error {
	privateKey, publicKey := []byte(conf.PrivateKey), []byte(conf.PublicKey)
	if conf.Format == "base64" {
		var err error
		privateKey, err = utils.B64Decode(conf.PrivateKey)
		if err != nil {
			return fmt.Errorf("descrypt private key error: %w", err)
		}
		publicKey, err = utils.B64Decode(conf.PublicKey)
		if err != nil {
			return fmt.Errorf("descrypt public key error: %w", err)
		}
	}
	var err error
	e.privateKey, err = parsePrivateKey(privateKey)
	if err != nil {
		return err
	}
	e.publicKey, err = parsePublicKey(publicKey)
	if err != nil {
		return err
	}
	e.requestSignHeader = conf.RequestSignHeader
	e.responseSignHeader = conf.ResponseSignHeader
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
