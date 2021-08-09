package apikey

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"apikey",
	"apikey_auth",
	"apikey-auth",
	"apikeyauth",
}

type apikey struct {
	id             string
	name           string
	driver         string
	hideCredential bool
	users          *apiKeyUsres
}

//Auth 鉴权处理
func (a *apikey) Auth(ctx *http_context.Context) error {
	authorizationType, has := ctx.Request().Header().Get(auth.AuthorizationType)
	if !has {
		return auth.ErrorInvalidType
	}
	// 判断是否要鉴权要求
	err := auth.CheckAuthorizationType(supportTypes, authorizationType)
	if err != nil {
		return err
	}
	authorization, err := a.getAuthValue(ctx)
	if err != nil {
		return err
	}
	for _, user := range a.users.users {
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
func (a *apikey) getAuthValue(ctx *http_context.Context) (string, error) {
	// 判断鉴权值是否在header

	if authorization, has := ctx.Request().Header().Get(auth.Authorization); has {
		if a.hideCredential {
			ctx.ProxyRequest().Header.Del(auth.Authorization)
		}
		return authorization, nil
	}

	// 判断鉴权值是否在query
	if authorization, has := ctx.Request().Query().Get("Apikey"); has {
		if a.hideCredential {
			ctx.ProxyRequest().URI().QueryArgs().Del("Apikey")
		}
		return authorization, nil
	}
	var authorization string
	contentType, _ := ctx.Request().Header().Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data") {
		formParams, err := ctx.BodyHandler().BodyForm()
		if err != nil {
			return "", err
		}
		authorization = formParams.Get("Apikey")
		if a.hideCredential {
			delete(formParams, "Apikey")
			ctx.BodyHandler().SetForm(formParams)
		}
	} else if strings.Contains(contentType, "application/json") {
		var body map[string]interface{}
		rawBody, err := ctx.BodyHandler().RawBody()
		if err != nil {
			return "", err
		}
		if err = json.Unmarshal(rawBody, &body); err != nil {
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

		if a.hideCredential {
			delete(body, "Apikey")
			newBody, err := json.Marshal(body)
			if err != nil {
				return "", err
			}
			ctx.BodyHandler().SetRaw(contentType, newBody)
		}

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
	a.users = &apiKeyUsres{
		users: cfg.User,
	}
	a.hideCredential = cfg.HideCredentials
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
