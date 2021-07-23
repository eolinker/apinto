package apikey

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"reflect"
	"strings"
	"time"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"apiKey",
	"apiKey_auth",
	"apiKey-auth",
	"apiKeyauth",
	"apikey",
	"apikey_auth",
	"apikey-auth",
	"apikeyauth",
	"Apikey",
	"Apikey_auth",
	"Apikey-auth",
	"Apikeyauth",
}

type apikey struct {
	id     string
	name   string
	driver string
	users  []User
}

//Auth 鉴权处理
func (a *apikey) Auth(ctx *http_context.Context) error {
	// 判断是否要鉴权要求
	err := auth.CheckAuthorizationType(supportTypes, ctx.Request().Headers().Get(auth.AuthorizationType))
	if err != nil {
		return err
	}
	authorization, err := getAuthValue(ctx)
	if err != nil {
		return err
	}
	for _, user := range a.users {
		if authorization == user.Apikey {
			if user.Expire == 0 || time.Now().Unix() < user.Expire {
				return nil
			}
			return auth.ErrorExpireUser
		}
	}
	return auth.ErrorInvalidUser

}

//TOfData 获取数据的类型
func TOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

//getAuthValue 获取Apikey值
func getAuthValue(ctx *http_context.Context) (string, error) {
	// 判断鉴权值是否在header
	authorization := ctx.Request().Headers().Get(auth.Authorization)
	if authorization != "" {
		return authorization, nil
	}
	authorization = ctx.Request().Headers().Get("Apikey")
	if authorization != "" {
		return authorization, nil
	}
	// 判断鉴权值是否在query
	authorization = ctx.Request().URL().Query().Get("Apikey")
	if authorization != "" {
		return authorization, nil
	}
	contentType := ctx.Request().Headers().Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "application/www-form-urlencoded") {
		formParams, err := ctx.Request().BodyForm()
		if err != nil {
			return "", err
		}
		authorization = formParams.Get("Apikey")
	} else if strings.Contains(contentType, "application/json") {
		var body map[string]interface{}
		rawbody, err := ctx.Request().RawBody()
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(rawbody, &body); err != nil {
			return "", err
		}
		if _, ok := body["Apikey"]; !ok {
			return "", errors.New("[apikey_auth] cannot find the Apikey in body")
		}
		if TOfData(body["Apikey"]) == reflect.String {
			authorization = body["Apikey"].(string)
		} else {
			return "", errors.New("[apikey_auth] Invalid data type for Apikey")
		}
	} else if strings.Contains(contentType, "multipart/form-data") {
		bodyform, err := ctx.Request().BodyForm()
		if err != nil {
			return "", err
		}
		authorization = bodyform.Get("Apikey")
	} else {
		return "", errors.New("[apikey_auth] Unsupported Content-Type")
	}
	if authorization != "" {
		return authorization, nil
	}
	return "", errors.New("[apikey_auth] cannot find the Apikey in query/body/header")
}

//Id 返回 worker ID
func (a *apikey) Id() string {
	return a.id
}

//Start
func (a *apikey) Start() error {
	return nil
}

//Reset 重新加载配置
func (a *apikey) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}
	a.users = cfg.User
	return nil
}

//Stop
func (a *apikey) Stop() error {
	return nil
}

//CheckSkill 技能检查
func (a *apikey) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}
