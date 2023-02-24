package grpc_to_http

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/jhump/protoreflect/desc"

	"github.com/valyala/fasthttp"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"

	"github.com/eolinker/eosc/log"

	"github.com/fullstorydev/grpcurl"

	"google.golang.org/grpc/metadata"

	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")

	defaultTimeout = 10 * time.Second
)

type complete struct {
	descSource grpcurl.DescriptorSource
	headers    map[string]string
	rawQuery   string
	path       string
	retry      int
	timeout    time.Duration
}

func newComplete(descSource grpcurl.DescriptorSource, conf *Config) *complete {
	timeout := defaultTimeout
	return &complete{
		descSource: descSource,
		timeout:    timeout,
	}
}

func (h *complete) Complete(org eocontext.EoContext) error {
	proxyTime := time.Now()
	ctx, err := grpc_context.Assert(org)
	if err != nil {
		return err
	}
	desc, err := h.descSource.FindSymbol(fmt.Sprintf("%s.%s", ctx.Proxy().Service(), ctx.Proxy().Method()))
	if err != nil {
		return err
	}
	methodDesc := desc.GetFile().FindService(ctx.Proxy().Service()).FindMethodByName(ctx.Proxy().Method())
	message := ctx.Proxy().Message(methodDesc.GetInputType())
	if err != nil {
		return err
	}
	body, err := message.Marshal()
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
	request := newRequest(ctx.Proxy().Headers(), body, h.headers, h.rawQuery)
	defer fasthttp.ReleaseRequest(request)
	var lastErr error
	timeOut := app.TimeOut()
	balance := ctx.GetBalance()
	for index := 0; index <= h.retry; index++ {

		if h.timeout > 0 && time.Now().Sub(proxyTime) > h.timeout {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			return status.Error(codes.NotFound, err.Error())
		}

		log.Debug("node: ", node.Addr())
		response := fasthttp.AcquireResponse()
		lastErr = fasthttp_client.ProxyTimeout(fmt.Sprintf("%s://%s", scheme, node.Addr()), request, response, timeOut)
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
	err := message.Unmarshal(response.Body())
	if err != nil {
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

func newRequest(headers metadata.MD, body []byte, additionalHeader map[string]string, rawQuery string) *fasthttp.Request {
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
	request.URI().SetQueryString(rawQuery)
	request.SetBody(body)
	return request
}
