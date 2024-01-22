package oauth2

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/eolinker/apinto/application/auth/oauth2"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	ResponseTypeCode  = "code"
	ResponseTypeToken = "token"
)

func (e *executor) Authorize(ctx http_context.IHttpContext, client oauth2.IClient, params url.Values) ([]byte, error) {
	responseType := params.Get("response_type")
	if responseType == "" || !((responseType == ResponseTypeCode && e.cfg.EnableAuthorizationCode) || (responseType == ResponseTypeToken && e.cfg.EnableImplicitGrant)) {
		return nil, fmt.Errorf("unsupported response type: %s,client id %s", responseType, client.ClientID())
	}

	scope := params.Get("scope")
	if scope == "" && e.cfg.MandatoryScope {
		return nil, fmt.Errorf("scope is required,client id %s", client.ClientID())
	}
	matchScope := false
	for _, s := range e.cfg.Scopes {
		if s == scope {
			matchScope = true
			break
		}
	}
	if len(e.cfg.Scopes) > 0 && !matchScope {
		return nil, fmt.Errorf("invalid scope,client id %s", client.ClientID())
	}

	redirectURI := params.Get("redirect_uri")
	matchRedirectUri := false
	for _, uri := range client.RedirectUrls() {
		if uri == redirectURI {
			matchRedirectUri = true
			break
		}
	}
	if len(client.RedirectUrls()) > 0 && !matchRedirectUri {
		return nil, fmt.Errorf("invalid redirect uri,client id %s", client.ClientID())
	}
	uri, err := url.Parse(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("parse redirect uri failed: %w,client id %s", err, client.ClientID())
	}

	list := e.cache.List()
	if len(list) < 1 {
		return nil, fmt.Errorf("redis cache is not available,client id %s", client.ClientID())
	}
	cache := list[0]
	query := url.Values{}
	switch responseType {
	case ResponseTypeCode:
		{
			// 授权码模式
			provisionKey := params.Get("provision_key")
			if provisionKey != e.cfg.ProvisionKey {
				return nil, fmt.Errorf("invalid provision key")
			}
			code := generateRandomString()
			redisKey := fmt.Sprintf("apinto:oauth2_codes:%s:%s", os.Getenv("cluster_id"), code)
			field := map[string]interface{}{
				"code":  code,
				"scope": scope,
			}
			_, err = cache.HMSetN(ctx.Context(), redisKey, field, 6*time.Minute).Result()
			if err != nil {
				return nil, fmt.Errorf("redis HMSet %s error: %w,client id %s", redisKey, err, client.ClientID())
			}
			query.Set("code", code)
		}
	case ResponseTypeToken:
		{
			token, err := generateToken(ctx.Context(), cache, client.ClientID(), e.cfg.TokenExpiration, e.cfg.RefreshTokenTtl, scope, false)
			if err != nil {
				return nil, fmt.Errorf("(%s)generate token error: %s", client.ClientID(), err.Error())
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
	return data, nil
}
