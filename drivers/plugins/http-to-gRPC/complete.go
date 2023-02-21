package http_to_grpc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/jhump/protoreflect/grpcreflect"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/eolinker/eosc/log"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"

	"github.com/fullstorydev/grpcurl"

	"google.golang.org/grpc/metadata"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
	options              = grpcurl.FormatOptions{
		AllowUnknownFields: true,
	}
	defaultTimeout = 10 * time.Second
)

type complete struct {
	format     grpcurl.Format
	descSource grpcurl.DescriptorSource
	timeout    time.Duration
	authority  string
	service    string
	method     string
	retry      int
	reflect    bool
}

func newComplete(descSource grpcurl.DescriptorSource, conf *Config) *complete {
	timeout := defaultTimeout
	return &complete{
		format:     grpcurl.Format(conf.Format),
		descSource: descSource,
		timeout:    timeout,
		authority:  conf.Authority,
		service:    conf.Service,
		method:     conf.Method,
		reflect:    conf.Reflect,
	}
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
	in := strings.NewReader(string(body))

	balance := ctx.GetBalance()
	app := ctx.GetApp()

	md := httpHeaderToMD(ctx.Proxy().Header().Headers())

	opts := genDialOpts(app.Scheme() == "https", h.authority)
	newCtx := metadata.NewOutgoingContext(ctx.Context(), md)
	symbol := fmt.Sprintf("%s/%s", h.service, h.method)

	var lastErr error
	var conn *grpc.ClientConn
	for i := h.retry + 1; i > 0; i-- {
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select node error: ", err)
			return err
		}
		conn, lastErr = dial(node.Addr(), h.timeout, opts...)
		if lastErr != nil {
			log.Error("dial error: ", lastErr)
			continue
		}
		descSource := h.descSource
		if h.reflect {
			refClient := grpcreflect.NewClientV1Alpha(newCtx, reflectpb.NewServerReflectionClient(conn))
			refSource := grpcurl.DescriptorSourceFromServer(newCtx, refClient)
			if h.descSource == nil {
				descSource = refSource
			} else {
				descSource = &compositeSource{reflection: refSource, file: h.descSource}
			}
		}

		rf, formatter, err := grpcurl.RequestParserAndFormatter(h.format, descSource, in, options)
		if err != nil {
			return fmt.Errorf("failed to construct request parser and formatter for %s", h.format)
		}
		response := NewResponse()
		handler := &grpcurl.DefaultEventHandler{
			Out:       response,
			Formatter: formatter,
		}
		err = grpcurl.InvokeRPC(newCtx, descSource, conn, symbol, []string{}, handler, rf.Next)
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
		ctx.Response().SetBody(response.Body())
		return nil
	}
	return lastErr
}

type StatusErr struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

var skipHeaderKey = map[string]struct{}{
	//"user-agent": struct{}{},
	//"host":           struct{}{},
	//"origin":     struct{}{},
	//"connection": struct{}{},
	//"content-length": struct{}{},
	//"accept": struct{}{},
	//"cookie": struct{}{},
}

func httpHeaderToMD(headers http.Header) metadata.MD {
	md := metadata.New(map[string]string{})
	for key, value := range headers {
		if strings.ToLower(key) == "user-agent" {
			md.Set("grpc-go", value...)
			continue
		}

		md.Set(key, value...)
	}
	md.Set("content-type", "application/grpc")
	md.Delete("connection")
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
