package grpc_to_http

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/entries/ctx_key"
	"github.com/eolinker/apinto/entries/router"
	"net/url"
	"strings"
	"time"

	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/jhump/protoreflect/desc"

	"github.com/valyala/fasthttp"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"

	"github.com/eolinker/eosc/log"

	"google.golang.org/grpc/metadata"

	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type complete struct {
	descriptor grpc_descriptor.IDescriptor
	headers    map[string]string
	rawQuery   string
	path       string
}

func newComplete(descriptor grpc_descriptor.IDescriptor, conf *Config) *complete {
	query := url.Values{}
	for key, value := range conf.Query {
		query.Set(key, value)
	}
	return &complete{
		descriptor: descriptor,
		rawQuery:   query.Encode(),
		path:       conf.Path,
		headers:    conf.Headers,
	}
}

func (h *complete) Complete(org eocontext.EoContext) error {
	proxyTime := time.Now()
	ctx, err := grpc_context.Assert(org)
	if err != nil {
		return err
	}

	retryValue := ctx.Value(ctx_key.CtxKeyRetry)
	retry, ok := retryValue.(int)
	if !ok {
		retry = router.DefaultRetry
	}

	timeoutValue := ctx.Value(ctx_key.CtxKeyTimeout)
	timeout, ok := timeoutValue.(time.Duration)
	if !ok {
		timeout = router.DefaultTimeout
	}

	descriptor, err := h.descriptor.Descriptor().FindSymbol(fmt.Sprintf("%s.%s", ctx.Proxy().Service(), ctx.Proxy().Method()))
	if err != nil {
		return err
	}
	methodDesc := descriptor.GetFile().FindService(ctx.Proxy().Service()).FindMethodByName(ctx.Proxy().Method())
	message := ctx.Proxy().Message(methodDesc.GetInputType())

	body, err := message.MarshalJSON()
	if err != nil {
		return err
	}

	app := ctx.GetApp()

	scheme := app.Scheme()
	switch strings.ToLower(app.Scheme()) {
	case "", "tcp":
		scheme = "http"
	case "tsl", "ssl", "https":
		scheme = "https"

	}
	path := h.path
	if path == "" {
		path = fmt.Sprintf("/%s/%s", ctx.Proxy().Service(), ctx.Proxy().Method())
	}
	request := newRequest(ctx.Proxy().Headers(), body, h.headers, path, h.rawQuery)
	defer fasthttp.ReleaseRequest(request)
	var lastErr error
	timeOut := app.TimeOut()
	balance := ctx.GetBalance()
	for index := 0; index <= retry; index++ {

		if timeout > 0 && time.Since(proxyTime) > timeout {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			return status.Error(codes.NotFound, err.Error())
		}

		request.URI()
		passHost, targetHost := ctx.GetUpstreamHostHandler().PassHost()
		switch passHost {
		case eocontext.PassHost:
			request.URI().SetHost(strings.Join(ctx.Proxy().Headers().Get(":authority"), ","))
		case eocontext.NodeHost:
			request.URI().SetHost(node.Addr())
		case eocontext.ReWriteHost:
			request.URI().SetHost(targetHost)
		}
		response := fasthttp.AcquireResponse()
		lastErr = fasthttp_client.ProxyTimeout(scheme, node, request, response, timeOut)
		if lastErr == nil {
			return newGRPCResponse(ctx, response, methodDesc)
		}
		log.Error("http upstream send error: ", lastErr)
	}

	return status.Error(codes.Internal, lastErr.Error())
}

func newGRPCResponse(ctx grpc_context.IGrpcContext, response *fasthttp.Response, methodDesc *desc.MethodDescriptor) error {
	defer fasthttp.ReleaseResponse(response)
	message := dynamic.NewMessage(methodDesc.GetOutputType())
	err := message.UnmarshalJSON(response.Body())
	if err != nil {
		log.Debug("body is: ", string(response.Body()))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	ctx.Response().Write(message)
	hs := strings.Split(response.Header.String(), "\r\n")
	for _, t := range hs {
		vs := strings.Split(t, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			ctx.Response().Headers().Set(vs[0], strings.TrimSpace(""))
			continue
		}
		ctx.Response().Headers().Set(vs[0], strings.TrimSpace(vs[1]))
	}

	return nil
}

func newRequest(headers metadata.MD, body []byte, additionalHeader map[string]string, path, rawQuery string) *fasthttp.Request {
	request := fasthttp.AcquireRequest()
	for key, value := range headers {
		if strings.ToLower(key) == "grpc-go" {
			key = "user-agent"
		}
		for _, v := range value {
			request.Header.Add(key, v)
		}
	}
	for key, value := range additionalHeader {
		request.Header.Add(key, value)
	}
	request.Header.Set("content-type", "application/json")
	request.URI().SetPath(path)
	request.URI().SetQueryString(rawQuery)
	request.SetBody(body)
	return request
}
