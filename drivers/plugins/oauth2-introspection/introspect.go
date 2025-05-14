package oauth2_introspection

import (
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type IntrospectionResponseBody struct {
	Active   bool   `json:"active"`
	ClientId string `json:"client_id"`
	Username string `json:"username"`
	Scope    string `json:"scope"`
	Sub      string `json:"sub"`
	Aud      string `json:"aud"`
	Iss      string `json:"iss"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	Nbf      int64  `json:"nbf"`
	Jti      string `json:"jti"`
}

func setAppLabel(ctx http_service.IHttpContext, t *IntrospectionResponseBody, consumerBy string, allowAnonymous bool) error {
	consumer := t.ClientId
	switch consumerBy {
	case "client_id":
	case "username":
		consumer = t.Username
	default:
		return fmt.Errorf("invalid consumer_by")
	}
	a, has := appManager.GetApp(consumer)
	if !has {
		if !allowAnonymous {
			return fmt.Errorf("consumer(%s) not found", consumer)
		}
		a = appManager.AnonymousApp()
		if a == nil {
			return fmt.Errorf("anonymous app not found")
		}
		ctx.Proxy().Header().SetHeader("X-Consumer-Anonymous", "true")
	}
	ctx.SetLabel("application_id", a.Id())
	ctx.SetLabel("application_name", a.Name())
	ctx.Proxy().Header().SetHeader("X-Consumer-ID", a.Id())
	ctx.Proxy().Header().SetHeader("X-Consumer-Username", a.Name())

	return nil
}

func verifyIntrospection(t *IntrospectionResponseBody, clientId string, scopes map[string]struct{}) error {
	if t.Active != true {
		return fmt.Errorf("token is not active")
	}
	if t.ClientId != clientId {
		return fmt.Errorf("invalid client_id")
	}

	now := time.Now()
	if t.Exp < now.Unix() {
		return fmt.Errorf("token is expired")
	}
	if t.Iat > now.Unix() {
		return fmt.Errorf("token is not yet active")
	}
	if len(scopes) > 0 {
		if _, ok := scopes[t.Scope]; !ok {
			return fmt.Errorf("invalid scope")
		}
	}

	return nil
}

func checkActive(t *IntrospectionResponseBody) bool {
	if t.Active != true {
		return false
	}
	now := time.Now()
	if t.Exp < now.Unix() {
		return false
	}
	if t.Iat > now.Unix() {
		return false
	}
	return true
}

func doIntrospectAccessToken(client *http.Client, endpoint string, clientId string, clientSecret string, token string) (*eosc.Base[IntrospectionResponseBody], error) {
	body := url.Values{}
	body.Set("token", token)
	body.Set("client_id", clientId)
	body.Set("client_secret", clientSecret)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(data))
	}
	t := new(eosc.Base[IntrospectionResponseBody])
	err = json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func retrieveAccessToken(ctx http_service.IHttpContext, tokenPosition string, tokenName string) string {
	token := ""
	switch tokenPosition {
	case positionHeader:
		token = ctx.Request().Header().GetHeader(tokenName)
		return strings.TrimPrefix(token, "Bearer ")
	case positionQuery:
		token = ctx.Request().URI().GetQuery(tokenName)
	case positionBody:
		if strings.Contains(ctx.Request().ContentType(), "application/x-www-form-urlencoded") || strings.Contains(ctx.Request().ContentType(), "multipart/form-data") {
			token = ctx.Request().Body().GetForm(tokenName)
		} else if strings.Contains(ctx.Request().ContentType(), "application/json") {
			body, _ := ctx.Request().Body().RawBody()
			if string(body) != "" {
				m := make(map[string]interface{})
				err := json.Unmarshal(body, &m)
				if err == nil {
					if v, ok := m[tokenName]; ok {
						token = fmt.Sprintf("%v", v)
					}
				} else {
					return ""
				}
			}
		}
	default:
		return ""
	}
	return strings.TrimPrefix(token, "Bearer ")
}
