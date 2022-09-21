package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/utils/config"
)

var (
	appSkill         string
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token is expired")
)

func init() {
	var t IApp
	appSkill = config.TypeNameOf(&t)
}

type IApp interface {
	Id() string
	Name() string
	Labels() map[string]string
	Disable() bool
	IAppExecutor
}

type IAppExecutor interface {
	Execute(ctx http_service.IHttpContext) error
}

func CheckSkill(skill string) bool {
	return skill == appSkill
}

type IAuth interface {
	ID() string
	Check(appID string, users []*BaseConfig) error
	Set(app IApp, users []*BaseConfig)
	Del(appID string)
	UserCount() int
	IAuthUser
}

type IAuthUser interface {
	Driver() string
	GetUser(ctx http_service.IHttpContext) (*UserInfo, bool)
}

type ITransformConfig interface {
	Config() interface{}
	SetType(typ reflect.Type) error
}

type BaseConfig struct {
	data []byte
	v    interface{}
}

func (a *BaseConfig) SetType(typ reflect.Type) error {
	v := reflect.New(typ.Elem()).Interface()
	err := json.Unmarshal(a.data, v)
	if err != nil {
		return fmt.Errorf("set type error: %v", err)
	}
	a.v = v
	return nil
}

func (a *BaseConfig) Config() interface{} {
	return a.v
}

func (a *BaseConfig) UnmarshalJSON(bytes []byte) error {
	a.data = bytes
	return nil
}

func (a *BaseConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.v)
}
