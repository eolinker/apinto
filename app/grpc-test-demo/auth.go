package main

import (
	"context"
	"liujian-test/grpc-test-demo/auth"
	"liujian-test/grpc-test-demo/auth/basic"
	"liujian-test/grpc-test-demo/auth/jwt"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

func initAuth() auth.AuthFunc {
	manager := auth.NewManager()
	manager.Register(basic.NewAuth(), jwt.NewAuth())
	return manager.GenAuthFunc()
}

func UnaryServerAuthInterceptor(authFunc auth.AuthFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error
		if overrideSrv, ok := info.Server.(grpc_auth.ServiceAuthFuncOverride); ok {
			newCtx, err = overrideSrv.AuthFuncOverride(ctx, info.FullMethod)
		} else {
			newCtx, err = authFunc(ctx, info.FullMethod)
		}
		if err != nil {
			//return nil, err
		}
		return handler(newCtx, req)
	}
}

func StreamServerAuthInterceptor(authFunc auth.AuthFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		_, err := authFunc(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
