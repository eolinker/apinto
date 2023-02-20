package grpc_proxy_rewrite

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"

	"google.golang.org/grpc/metadata"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type HttpComplete struct {
	refClient  *grpcreflect.Client
	fileSource grpcurl.DescriptorSource
	tls        bool
}

func NewHttpComplete(retry int, timeOut time.Duration) *HttpComplete {
	return &HttpComplete{}
}

func (h *HttpComplete) Complete(org eocontext.EoContext) error {

	ctx, err := http_context.Assert(org)
	if err != nil {
		return err
	}
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
	return md
}

func dial(ctx context.Context, target string) *grpc.ClientConn {
	dialTime := 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, dialTime)
	defer cancel()
	var opts []grpc.DialOption

	var creds credentials.TransportCredentials
	if !*plaintext {
		tlsConf, err := grpcurl.ClientTLSConfig(*insecure, *cacert, *cert, *key)
		if err != nil {
			fail(err, "Failed to create TLS config")
		}

		sslKeylogFile := os.Getenv("SSLKEYLOGFILE")
		if sslKeylogFile != "" {
			w, err := os.OpenFile(sslKeylogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				fail(err, "Could not open SSLKEYLOGFILE %s", sslKeylogFile)
			}
			tlsConf.KeyLogWriter = w
		}

		creds = credentials.NewTLS(tlsConf)

		// can use either -servername or -authority; but not both
		if *serverName != "" && *authority != "" {
			if *serverName == *authority {
				warn("Both -servername and -authority are present; prefer only -authority.")
			} else {
				fail(nil, "Cannot specify different values for -servername and -authority.")
			}
		}
		overrideName := *serverName
		if overrideName == "" {
			overrideName = *authority
		}

		if overrideName != "" {
			opts = append(opts, grpc.WithAuthority(overrideName))
		}
	} else if *authority != "" {
		opts = append(opts, grpc.WithAuthority(*authority))
	}

	grpcurlUA := "grpcurl/" + version
	if version == noVersion {
		grpcurlUA = "grpcurl/dev-build (no version set)"
	}
	if *userAgent != "" {
		grpcurlUA = *userAgent + " " + grpcurlUA
	}
	opts = append(opts, grpc.WithUserAgent(grpcurlUA))

	network := "tcp"
	if isUnixSocket != nil && isUnixSocket() {
		network = "unix"
	}
	cc, err := grpcurl.BlockingDial(ctx, network, target, creds, opts...)
	if err != nil {
		fail(err, "Failed to dial target host %q", target)
	}
	return cc
}
