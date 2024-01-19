package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/resources"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	ResponseTypeCode  = "code"
	ResponseTypeToken = "token"
)

func NewAuthorizeHandler() *AuthorizeHandler {
	return &AuthorizeHandler{}
}

type AuthorizeHandler struct {
	cache scope_manager.IProxyOutput[resources.ICache]
	once  sync.Once
}

func (a *AuthorizeHandler) Handle(ctx http_context.IHttpContext, client *Client, params url.Values) {
	responseType := params.Get("response_type")
	if responseType == "" || !((responseType == ResponseTypeCode && client.EnableAuthorizationCode) || (responseType == ResponseTypeToken && client.EnableImplicitGrant)) {
		ctx.Response().SetBody([]byte(fmt.Sprintf("unsupported response type: %s,client id is %s", responseType, client.ClientId)))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}

	scope := params.Get("scope")
	if scope == "" && client.MandatoryScope {
		ctx.Response().SetBody([]byte("scope is required, client id is " + client.ClientId))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}
	matchScope := false
	for _, s := range client.Scopes {
		if s == scope {
			matchScope = true
			break
		}
	}
	if !matchScope {
		ctx.Response().SetBody([]byte("invalid scope, client id is " + client.ClientId))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}

	redirectURI := params.Get("redirect_uri")
	if redirectURI == "" {
		ctx.Response().SetBody([]byte("redirect uri is required, client id is " + client.ClientId))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}

	matchRedirectUri := false
	for _, uri := range client.RedirectUrls {
		if uri == redirectURI {
			matchRedirectUri = true
			break
		}
	}
	if !matchRedirectUri {
		ctx.Response().SetBody([]byte("invalid redirect uri, client id is " + client.ClientId))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}
	uri, err := url.Parse(redirectURI)
	if err != nil {
		ctx.Response().SetBody([]byte("invalid redirect uri, client id is " + client.ClientId))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}
	a.once.Do(func() {
		a.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	list := a.cache.List()
	if len(list) < 1 {
		ctx.Response().SetBody([]byte("redis cache is not available"))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}
	cache := list[0]
	query := url.Values{}
	switch responseType {
	case ResponseTypeCode:
		{
			// 授权码模式
			provisionKey := params.Get("provision_key")
			if provisionKey != client.ProvisionKey {
				ctx.Response().SetBody([]byte("invalid provision key, client id is " + client.ClientId))
				ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
				return
			}
			code := generateRandomString()
			redisKey := fmt.Sprintf("apinto:oauth2_codes:%s:%s", os.Getenv("cluster_id"), code)
			field := map[string]interface{}{
				"code":  code,
				"scope": scope,
			}
			_, err = cache.HMSetN(ctx.Context(), redisKey, field, 6*time.Minute).Result()
			if err != nil {
				ctx.Response().SetBody([]byte(fmt.Sprintf("(%s)redis HMSet %s error: %s", client.ClientId, redisKey, err.Error())))
				ctx.Response().SetStatus(http.StatusInternalServerError, "server error")
				return
			}
			query.Set("code", code)
		}
	case ResponseTypeToken:
		{
			token, err := generateToken(ctx.Context(), cache, client.ClientId, client.TokenExpiration, client.RefreshTokenTTL, scope, false)
			if err != nil {
				ctx.Response().SetBody([]byte(fmt.Sprintf("(%s)generate token error: %s", client.ClientId, err.Error())))
				ctx.Response().SetStatus(http.StatusInternalServerError, "server error")
				return
			}
			query.Set("access_token", token.AccessToken)
			query.Set("token_type", "bearer")
			query.Set("expires_in", strconv.Itoa(token.ExpiresIn))
		}
	}

	state := params.Get("state")
	if state != "" {
		query.Set("state", state)
	}
	data, _ := json.Marshal(map[string]interface{}{
		"redirect_uri": fmt.Sprintf("%s?%s", uri.String(), query.Encode()),
	})
	ctx.Response().SetBody(data)
	ctx.Response().SetStatus(http.StatusOK, "OK")
	return
}
