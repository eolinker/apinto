package http_to_grpc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/eolinker/apinto/entries/ctx_key"
	"github.com/eolinker/apinto/entries/router"
	"net/http"
	"strings"
	"time"

	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"

	"google.golang.org/grpc/codes"

	"github.com/jhump/protoreflect/grpcreflect"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/eolinker/eosc/log"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"

	"github.com/fullstorydev/grpcurl"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc/eocontext"
)

var (
	options = grpcurl.FormatOptions{
		AllowUnknownFields: true,
	}
)

type complete struct {
	format     grpcurl.Format
	descriptor grpc_descriptor.IDescriptor
	authority  string
	service    string
	method     string
	headers    map[string]string
	reflect    bool
}

func newComplete(descriptor grpc_descriptor.IDescriptor, conf *Config) *complete {
	return &complete{
		format:     grpcurl.Format(conf.Format),
		descriptor: descriptor,
		authority:  conf.Authority,
		service:    conf.Service,
		method:     conf.Method,
		reflect:    conf.Reflect,
		headers:    conf.Headers,
	}
}

func getSymbol(path string, service string, method string) string {
	ps := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if service == "" {
		service = ps[0]
	}
	if method == "" {
		if len(ps) > 1 {
			method = ps[1]
		}
	}
	return fmt.Sprintf("%s/%s", service, method)

}

func (h *complete) Complete(org eocontext.EoContext) error {

	ctx, err := http_context.Assert(org)
	if err != nil {
		return err
	}
	body, err := ctx.Proxy().Body().RawBody()
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

	in := strings.NewReader(string(body))

	balance := ctx.GetBalance()
	app := ctx.GetApp()

	md := httpHeaderToMD(ctx.Proxy().Header().Headers(), h.headers)
	newCtx := ctx.Context()
	opts := genDialOpts(app.Scheme() == "https", h.authority)

	symbol := getSymbol(ctx.Proxy().URI().Path(), h.service, h.method)
	var lastErr error
	var conn *grpc.ClientConn
	for i := retry + 1; i > 0; i-- {
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select node error: ", err)
			return err
		}
		conn, lastErr = dial(node.Addr(), timeout, opts...)
		if lastErr != nil {
			log.Error("dial error: ", lastErr)
			continue
		}
		var descSource grpcurl.DescriptorSource
		if h.reflect {
			refClient := grpcreflect.NewClientV1Alpha(newCtx, reflectpb.NewServerReflectionClient(conn))
			refSource := grpcurl.DescriptorSourceFromServer(newCtx, refClient)
			if descSource == nil {
				descSource = refSource
			} else {
				descSource = &compositeSource{reflection: refSource, file: descSource}
			}
		} else {
			descSource = h.descriptor.Descriptor()
		}

		rf, formatter, err := grpcurl.RequestParserAndFormatter(h.format, descSource, in, options)
		if err != nil {
			return fmt.Errorf("failed to construct request parser and formatter for %s", h.format)
		}
		response := NewResponse()
		handler := &grpcurl.DefaultEventHandler{
			VerbosityLevel: 2,
			Out:            response,
			Formatter:      formatter,
		}
		err = grpcurl.InvokeRPC(newCtx, descSource, conn, symbol, md, handler, rf.Next)
		if err != nil {
			if errStatus, ok := status.FromError(err); ok {
				data, _ := json.Marshal(StatusErr{
					Code: fmt.Sprintf("%s", errStatus.Code()),
					Msg:  errStatus.Message(),
				})
				ctx.Response().SetBody(data)
				return err
			}
			err = fmt.Errorf("error invoking method %s", symbol)
			data, _ := json.Marshal(StatusErr{
				Code: fmt.Sprintf("%s", codes.Unavailable),
				Msg:  err.Error(),
			})

			ctx.Response().SetBody(data)
			return err
		}
		for key, value := range response.Header() {
			ctx.Response().SetHeader(key, value)
		}
		ctx.Response().SetHeader("content-type", "application/json")
		ctx.Response().SetBody(response.Body())
		return nil
	}
	return lastErr
}

type StatusErr struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func httpHeaderToMD(headers http.Header, additionalHeader map[string]string) []string {
	headers.Set("content-type", "application/grpc")
	headers.Del("connection")
	md := make([]string, len(headers)+len(additionalHeader))
	//md := metadata.New(map[string]string{})
	for key, value := range headers {
		if strings.ToLower(key) == "user-agent" {
			for _, v := range value {
				md = append(md, fmt.Sprintf("%s: %s", key, v))
			}
			continue
		}
		for _, v := range value {
			md = append(md, fmt.Sprintf("%s: %s", key, v))
		}
	}
	for key, value := range additionalHeader {
		md = append(md, fmt.Sprintf("%s: %s", key, value))
	}
	return md
}

func genDialOpts(isTLS bool, authority string) []grpc.DialOption {
	var opts []grpc.DialOption
	if isTLS {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if authority != "" {
		opts = append(opts, grpc.WithAuthority(authority))
	}

	return opts
}

func dial(target string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cc, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
