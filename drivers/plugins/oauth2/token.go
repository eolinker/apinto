package oauth2

import (
	"context"
	"crypto/md5"
	"crypto/rand"
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

	"github.com/eolinker/apinto/utils"

	"github.com/eolinker/apinto/application/auth/oauth2"

	"github.com/eolinker/eosc/log"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/resources"
)

type TokenResponse struct {
	Total int          `json:"total"`
	Data  []*TokenData `json:"data"`
}

type TokenData struct {
	AuthenticatedUserid interface{} `json:"authenticated_userid"`
	Credential          struct {
		Id string `json:"id"`
	} `json:"credential"`
	AccessToken  string      `json:"access_token"`
	Service      interface{} `json:"service"`
	CreatedAt    int64       `json:"created_at"`
	RefreshToken interface{} `json:"refresh_token"`
	Scope        interface{} `json:"scope"`
	Ttl          int         `json:"ttl"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	ClientID     string      `json:"client_id"`
}

func NewTokenHandler() *TokenHandler {
	h := &TokenHandler{}

	return h
}

type TokenHandler struct {
	cache scope_manager.IProxyOutput[resources.ICache]
	once  sync.Once
}

func (t *TokenHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	t.once.Do(func() {
		t.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	list := t.cache.List()
	if len(list) < 1 {
		writer.WriteHeader(http.StatusOK)
		writer.Write(newError(http.StatusForbidden, "redis cache not found"))
		return
	}
	cache := list[0]
	switch request.Method {
	case http.MethodPost:
		// 创建token
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Write(newError(http.StatusForbidden, err.Error()))
			return
		}
		var resp TokenResponse
		err = json.Unmarshal(body, &resp)
		if err != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Write(newError(-1, err.Error()))
			return
		}
		for _, token := range resp.Data {
			createAt := time.UnixMilli(token.CreatedAt)
			if createAt.Add(time.Duration(token.ExpiresIn) * time.Second).Before(time.Now()) {
				// 过期
				continue
			}
			redisKey := fmt.Sprintf("apinto:oauth2_access_tokens:%s:%s", os.Getenv("cluster_id"), token.AccessToken)
			// 保存token
			cache.HMSetN(context.Background(), redisKey, map[string]interface{}{
				"access_token":  token.AccessToken,
				"scope":         token.Scope,
				"expires_in":    token.ExpiresIn,
				"create_at":     token.CreatedAt,
				"refresh_token": token.RefreshToken,
				"client_id":     token.Credential.Id,
			}, time.Duration(token.ExpiresIn)*time.Second)
		}
		byteBody, _ := json.Marshal(map[string]interface{}{
			"code": 0,
		})
		writer.WriteHeader(http.StatusOK)
		writer.Write(byteBody)
		return
	case http.MethodGet:
		// 获取tokens
		tokenKeys, err := cache.Keys(context.Background(), fmt.Sprintf("apinto:oauth2_access_tokens:%s:*", os.Getenv("cluster_id"))).Result()
		if err != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Write(newError(-1, err.Error()))
			return
		}
		var tokens []*TokenData
		for _, key := range tokenKeys {
			token, err := getTokenByRedis(cache, key)
			if err != nil {
				log.Errorf("get token error: %s", err.Error())
				continue
			}
			tokens = append(tokens, token)
		}
		data, err := json.Marshal(TokenResponse{
			Total: len(tokens),
			Data:  tokens,
		})
		if err != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Write(newError(-1, err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(data)
		return
	}
}

func getTokenByRedis(cache resources.ICache, redisKey string) (*TokenData, error) {
	var accessToken, scope, refreshToken, clientId, createdAt, expiresIn string
	result, err := cache.HMGet(context.Background(), redisKey, "access_token", "scope", "expires_in", "create_at", "refresh_token", "client_id").Result()
	if err != nil {
		return nil, err
	}
	_, err = utils.Scan(result, &accessToken, &scope, &expiresIn, &createdAt, &refreshToken, &clientId)
	if err != nil {
		return nil, err
	}
	expiresInInt, err := strconv.Atoi(expiresIn)
	if err != nil {
		return nil, err
	}
	createdAtInt, err := strconv.ParseInt(createdAt, 10, 64)
	if err != nil {
		return nil, err
	}
	return &TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Scope:        scope,
		ExpiresIn:    expiresInInt,
		CreatedAt:    createdAtInt,
		ClientID:     clientId,
	}, nil
}

func newError(code int, msg string) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"code": code,
		"err":  msg,
	})
	return body
}

func (e *executor) Token(ctx http_service.IHttpContext, client oauth2.IClient, params url.Values) ([]byte, error) {

	grantType := params.Get("grant_type")
	clientSecret := params.Get("client_secret")
	state := params.Get("state")
	if grantType == "" || !((grantType == GrantAuthorizationCode && e.cfg.EnableAuthorizationCode) || (grantType == GrantClientCredentials && e.cfg.EnableClientCredentials) || grantType == GrantRefreshToken) {
		return nil, fmt.Errorf("unsupported grant type: %s,client id %s", grantType, client.ClientID())
	}
	err := client.MatchSecret(clientSecret)
	if err != nil {
		return nil, fmt.Errorf("client secret is not match: %s,client id %s", err.Error(), client.ClientID())
	}

	type Response struct {
		*Token
		State string `json:"state,omitempty"`
	}

	list := e.cache.List()
	if len(list) < 1 {
		return nil, fmt.Errorf("redis cache is not found,client id %s", client.ClientID())
	}

	cache := list[0]
	switch grantType {
	case GrantRefreshToken:
		refreshToken := params.Get("refresh_token")
		if refreshToken == "" {
			return nil, fmt.Errorf("refresh token is required,client id %s", client.ClientID())
		}

		redisKey := fmt.Sprintf("apinto:oauth2_refresh_tokens:%s:%s", os.Getenv("cluster_id"), refreshToken)

		result, err := cache.HMGet(ctx.Context(), redisKey, "refresh_token", "access_token").Result()
		if err != nil {
			return nil, fmt.Errorf("fail to get refresh token,client id %s", client.ClientID())
		}
		var refreshTokenStr, accessTokenStr string
		_, err = utils.Scan(result, &refreshTokenStr, &accessTokenStr)
		if err != nil {
			return nil, fmt.Errorf("scan refresh token error: %s,client id %s", err.Error(), client.ClientID())
		}
		if refreshTokenStr != refreshToken {
			return nil, fmt.Errorf("invalid refresh token,client id %s", client.ClientID())
		}
		token, err := generateToken(ctx.Context(), cache, client.ClientID(), e.cfg.TokenExpiration, e.cfg.RefreshTokenTtl, "", !e.cfg.ReuseRefreshToken)
		if !e.cfg.PersistentRefreshToken {
			// 不持久化refresh token
			accessTokenRedisKey := fmt.Sprintf("apinto:oauth2_access_tokens:%s:%s", os.Getenv("cluster_id"), accessTokenStr)
			cache.Del(ctx.Context(), accessTokenRedisKey)
		}
		if e.cfg.ReuseRefreshToken {
			// 重用refresh token
			token.AccessToken = accessTokenStr
			cache.HMSetN(ctx.Context(), redisKey, map[string]interface{}{
				"access_token": token.AccessToken,
			}, 0)
		} else {
			cache.Del(ctx.Context(), redisKey)
		}
		response := &Response{
			Token: token,
			State: state,
		}
		data, _ := json.Marshal(response)
		return data, nil
	case GrantAuthorizationCode:
		code := params.Get("code")
		if code == "" {
			return nil, fmt.Errorf("code is required,client id %s", client.ClientID())
		}
		redisKey := fmt.Sprintf("apinto:oauth2_codes:%s:%s", os.Getenv("cluster_id"), code)
		result, err := cache.HMGet(ctx.Context(), redisKey, "code", "scope").Result()
		if err != nil {
			return nil, fmt.Errorf("fail to get code: %s,client id %s", err.Error(), client.ClientID())
		}
		// 删除旧授权码
		cache.Del(ctx.Context(), redisKey)
		var codeStr, scope string
		_, err = utils.Scan(result, &codeStr, &scope)
		if err != nil {
			return nil, fmt.Errorf("scan code error: %s,client id %s", err.Error(), client.ClientID())
		}
		if codeStr != code {
			return nil, fmt.Errorf("invalid code,client id %s", client.ClientID())
		}

		token, err := generateToken(ctx.Context(), cache, client.ClientID(), e.cfg.TokenExpiration, e.cfg.RefreshTokenTtl, scope, true)
		if err != nil {
			return nil, fmt.Errorf("(%s)generate token error: %s", client.ClientID(), err.Error())
		}
		response := &Response{
			Token: token,
			State: state,
		}
		data, _ := json.Marshal(response)
		return data, nil
	case GrantClientCredentials:
		// 生成token
		token, err := generateToken(ctx.Context(), cache, client.ClientID(), e.cfg.TokenExpiration, e.cfg.RefreshTokenTtl, "", false)
		if err != nil {
			return nil, fmt.Errorf("(%s)generate token error: %s", client.ClientID(), err.Error())
		}
		response := &Response{
			Token: token,
			State: state,
		}
		data, _ := json.Marshal(response)
		return data, nil
	}
	return nil, fmt.Errorf("unsupported grant type: %s,client id %s", grantType, client.ClientID())
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

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
