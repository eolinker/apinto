package oauth2

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"golang.org/x/crypto/pbkdf2"

	"github.com/eolinker/apinto/resources"
)

func NewTokenHandler() *TokenHandler {
	return &TokenHandler{}
}

type TokenHandler struct {
	cache scope_manager.IProxyOutput[resources.ICache]
	once  sync.Once
}

func (t *TokenHandler) Handle(ctx http_service.IHttpContext, client *Client, params url.Values) {

	grantType := params.Get("grant_type")
	clientSecret := params.Get("client_secret")
	state := params.Get("state")
	if grantType == "" || !((grantType == GrantAuthorizationCode && client.EnableAuthorizationCode) || (grantType == GrantClientCredentials && client.EnableClientCredentials)) {
		ctx.Response().SetBody([]byte(fmt.Sprintf("unsupported grant type: %s,client id is %s", grantType, client.ClientId)))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}

	if client.HashSecret {
		// 密钥经过加密
		salt, _ := base64.RawStdEncoding.DecodeString(client.hashRule.salt)
		secret := pbkdf2.Key([]byte(clientSecret), salt, client.hashRule.iterations, client.hashRule.length, sha512.New)
		clientSecret = base64.RawStdEncoding.EncodeToString(secret)
	}

	if clientSecret != client.hashRule.value {
		ctx.Response().SetBody([]byte(fmt.Sprintf("fail to match secret,now: %s,hope: %s,client id is %s", clientSecret, client.hashRule.value, client.ClientId)))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}
	type Response struct {
		*Token
		State string `json:"state,omitempty"`
	}
	t.once.Do(func() {
		t.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	list := t.cache.List()
	if len(list) < 1 {
		ctx.Response().SetBody([]byte("redis cache is not found"))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return
	}

	cache := list[0]
	switch grantType {
	case GrantRefreshToken:
		refreshToken := params.Get("refresh_token")
		if refreshToken == "" {
			ctx.Response().SetBody([]byte("refresh token is required, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}

		redisKey := fmt.Sprintf("apinto:oauth2_refresh_tokens:%s:%s", os.Getenv("cluster_id"), refreshToken)

		result, err := cache.HMGet(ctx.Context(), redisKey, "refresh_token", "access_token").Result()
		if err != nil {
			ctx.Response().SetBody([]byte("fail to get refresh token, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		var refreshTokenStr, accessTokenStr string
		_, err = Scan(result, &refreshTokenStr, &accessTokenStr)
		if err != nil {
			ctx.Response().SetBody([]byte("invalid refresh token, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		if refreshTokenStr != refreshToken {
			ctx.Response().SetBody([]byte("invalid refresh token, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		token, err := generateToken(ctx.Context(), cache, client.ClientId, client.TokenExpiration, client.RefreshTokenTTL, "", !client.ReuseRefreshToken)
		if !client.PersistentRefreshToken {
			// 不持久化refresh token
			accessTokenRedisKey := fmt.Sprintf("apinto:oauth2_access_tokens:%s:%s", os.Getenv("cluster_id"), accessTokenStr)
			cache.Del(ctx.Context(), accessTokenRedisKey)
		}
		if client.ReuseRefreshToken {
			// 重用refresh token
			token.AccessToken = accessTokenStr
			cache.HMSetN(ctx.Context(), redisKey, map[string]interface{}{
				"access_token": token.AccessToken,
			}, 0)
		} else {
			cache.Del(ctx.Context(), redisKey)
		}
	case GrantAuthorizationCode:
		code := params.Get("code")
		if code == "" {
			ctx.Response().SetBody([]byte("code is required, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		redisKey := fmt.Sprintf("apinto:oauth2_codes:%s:%s", os.Getenv("cluster_id"), code)
		result, err := cache.HMGet(ctx.Context(), redisKey, "code", "scope").Result()
		if err != nil {
			ctx.Response().SetBody([]byte("fail to get code, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		var codeStr, scope string
		_, err = Scan(result, &codeStr, &scope)
		if err != nil {
			ctx.Response().SetBody([]byte("invalid code, client id is " + client.ClientId))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		token, err := generateToken(ctx.Context(), cache, client.ClientId, client.TokenExpiration, client.RefreshTokenTTL, scope, true)
		if err != nil {
			ctx.Response().SetBody([]byte(fmt.Sprintf("(%s)generate token error: %s", client.ClientId, err.Error())))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		response := &Response{
			Token: token,
			State: state,
		}
		data, _ := json.Marshal(response)
		ctx.Response().SetBody(data)
		ctx.Response().SetStatus(http.StatusOK, "ok")
		return
	case GrantClientCredentials:
		// 生成token
		token, err := generateToken(ctx.Context(), cache, client.ClientId, client.TokenExpiration, client.RefreshTokenTTL, "", false)
		if err != nil {
			ctx.Response().SetBody([]byte(fmt.Sprintf("(%s)generate token error: %s", client.ClientId, err.Error())))
			ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
			return
		}
		response := &Response{
			Token: token,
			State: state,
		}
		data, _ := json.Marshal(response)
		ctx.Response().SetBody(data)
		ctx.Response().SetStatus(http.StatusOK, "ok")
		return
	}
}

func generateRandomString() string {
	b := make([]byte, 40)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	baseRes := base64.StdEncoding.EncodeToString(b)
	h := md5.New()
	h.Write([]byte(baseRes))
	res := hex.EncodeToString(h.Sum(nil))
	return res
}

func retrieveParameters(ctx http_service.IHttpContext) url.Values {
	params := url.Values{}
	queries, _ := url.ParseQuery(ctx.Request().URI().RawQuery())
	for k, v := range queries {
		params.Set(k, v[0])
	}
	if strings.Contains(ctx.Request().ContentType(), "application/x-www-form-urlencoded") {
		body, _ := ctx.Request().Body().BodyForm()
		for k, v := range body {
			params.Set(k, v[0])
		}
	} else if strings.Contains(ctx.Request().ContentType(), "application/json") {
		var body map[string]string
		rawBody, _ := ctx.Request().Body().RawBody()
		json.Unmarshal(rawBody, &body)
		for k, v := range body {
			params.Set(k, v)
		}
	}
	return params
}

func generateToken(ctx context.Context, cache resources.ICache, clientID string, tokenExpired int, refreshTokenTTL int, scope string, isRefresh bool) (*Token, error) {
	// 简化模式
	accessToken := generateRandomString()
	if tokenExpired <= 0 {
		tokenExpired = 7200
	}
	if refreshTokenTTL <= 0 {
		refreshTokenTTL = 1209600
	}
	refreshToken := ""
	if isRefresh {
		refreshToken = generateRandomString()
	}

	redisKey := fmt.Sprintf("apinto:oauth2_access_tokens:%s:%s", os.Getenv("cluster_id"), accessToken)
	now := time.Now()
	fields := map[string]interface{}{
		"client_id":     clientID,
		"expires_in":    tokenExpired,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"create_at":     now.UnixMilli(),
		"scope":         scope,
	}
	_, err := cache.HMSetN(ctx, redisKey, fields, time.Duration(tokenExpired)*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("(%s)redis HMSet %s error: %s", clientID, redisKey, err.Error())
	}
	if isRefresh {
		redisKey = fmt.Sprintf("apinto:oauth2_refresh_tokens:%s:%s", os.Getenv("cluster_id"), refreshToken)

		_, err = cache.HMSetN(ctx, redisKey, fields, time.Duration(refreshTokenTTL)*time.Second).Result()
		if err != nil {
			return nil, fmt.Errorf("(%s)redis HMSet %s error: %s", clientID, redisKey, err.Error())
		}
	}
	return &Token{
		TokenType:    "bearer",
		ExpiresIn:    tokenExpired,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Scope:        scope,
	}, nil
}

func validToken(ctx context.Context, cache resources.ICache, token string) (string, error) {
	redisKey := fmt.Sprintf("apinto:oauth2_access_tokens:%s:%s", os.Getenv("cluster_id"), token)
	result, err := cache.HMGet(ctx, redisKey, "client_id", "access_token", "create_at", "expires_in").Result()
	if err != nil {
		return "", fmt.Errorf("redis HMGet %s error: %s", redisKey, err.Error())
	}
	var clientID, accessToken, createAt, expiresInStr string
	_, err = Scan(result, &clientID, &accessToken, &createAt, &expiresInStr)
	if err != nil {
		return "", fmt.Errorf("scan redis result error: %s", err.Error())
	}
	createAtTime, _ := strconv.ParseInt(createAt, 10, 64)
	expiresIn, _ := strconv.ParseInt(expiresInStr, 10, 64)
	createTime := time.UnixMilli(createAtTime)
	if time.Now().After(createTime.Add(time.Duration(expiresIn) * time.Second)) {
		// token过期
		return "", fmt.Errorf("token expired")
	}
	if accessToken != token {
		return "", fmt.Errorf("invalid token")
	}
	return clientID, nil
}

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
