package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"liujian-test/grpc-test-demo/common/flag"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/metadata"
)

var defaultManager = NewManager()

type IAuthFunc interface {
	Names() []string
	Auth(data map[string][]string) (string, error)
}

type Manager struct {
	auth map[string]IAuthFunc
}

func NewManager() *Manager {
	return &Manager{auth: map[string]IAuthFunc{}}
}

func (m *Manager) Register(authFunc ...IAuthFunc) {
	for _, f := range authFunc {
		for _, n := range f.Names() {
			m.auth[n] = f
		}
	}
}

func (m *Manager) GenAuthFunc() AuthFunc {
	return func(ctx context.Context, service string) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
		}
		authorizationType, ok := md["authorization-type"]
		if !ok {
			return nil, grpc.Errorf(codes.InvalidArgument, "authorization-type非法")
		}
		f, ok := m.auth[strings.ToLower(authorizationType[0])]
		if !ok {
			return nil, grpc.Errorf(codes.InvalidArgument, "authorization-type非法")
		}
		info := strings.Split(service, "/")
		if len(info) < 2 {
			return nil, grpc.Errorf(codes.NotFound, "请求路径不存在")
		}
		username, err := f.Auth(md)
		if err != nil {
			return nil, err
		}
		if !checkAccess(username, info[1]) {
			return ctx, errors.New("no permission")
		}
		return ctx, nil
	}
}

func checkAccess(username string, service string) bool {
	uri := fmt.Sprintf("%s/api/user", flag.ConfigAddress)
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	type response struct {
		Data map[string]map[string][]interface{} `json:"data"`
	}
	data := new(response)
	err = json.Unmarshal(body, data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if v, ok := data.Data["user"]; ok {
		if t, ok := v[username]; ok {
			for _, s := range t {
				if s == service {
					return true
				}
			}
		}
	}

	return false
}

type AuthFunc func(ctx context.Context, service string) (context.Context, error)
