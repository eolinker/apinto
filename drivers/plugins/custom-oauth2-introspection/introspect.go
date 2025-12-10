package custom_oauth2_introspection

import (
	"encoding/json"
	"fmt"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/oj"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	client = http.Client{
		Timeout: 5 * time.Second,
	}
)

type BaseResponse struct {
	Status int
	Header http.Header
	Body   []byte
}

type IntrospectResponse struct {
	Token     string
	ExpiredIn int
}

type IntrospectClient struct {
	Endpoint        string
	Method          string
	Header          map[string]string
	Body            map[string]string
	Query           map[string]string
	ContentType     string
	ExtractResponse *ExtractResponse
}

func retrieveToken(param *ExtractParam, header http.Header, body interface{}) (string, error) {
	tokenVal, err := retrieveResponseParam(param, header, body)
	if err != nil {
		return "", err
	}
	var token string
	switch tv := tokenVal.(type) {
	case string:
		token = tv
	default:
		b, _ := json.Marshal(tv)
		token = string(b)
	}
	if token == "" {
		return "", fmt.Errorf("empty access_token")
	}
	return token, nil
}

func retrieveExpireIn(param *ExtractParam, header http.Header, body interface{}) (int, error) {
	if param == nil {
		return 0, nil
	}
	expiredVal, err := retrieveResponseParam(param, header, body)
	if err != nil {
		return 0, err
	}
	var expiredIn int
	switch v := expiredVal.(type) {
	case string:
		if v == "" {
			expiredIn = 0
		} else {
			if n, err := strconv.Atoi(v); err == nil {
				expiredIn = n
			} else {
				expiredIn = 0
			}
		}
	case int:
		expiredIn = v
	case int64:
		expiredIn = int(v)
	case float64:
		expiredIn = int(v)
	default:
		expiredIn = 0
	}
	return expiredIn, nil
}

func (i *IntrospectClient) Do(ctx http_service.IHttpContext) (*IntrospectResponse, error) {
	resp, err := i.Send(ctx)
	if err != nil {
		return nil, err
	}
	data, err := oj.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	token, err := retrieveToken(i.ExtractResponse.AccessToken, resp.Header, data)
	if err != nil {
		return nil, err
	}
	expiredIn, err := retrieveExpireIn(i.ExtractResponse.ExpiredIn, resp.Header, data)
	if err != nil {
		return nil, err
	}

	return &IntrospectResponse{Token: token, ExpiredIn: expiredIn}, nil
}

func retrieveResponseParam(param *ExtractParam, header http.Header, body interface{}) (interface{}, error) {
	switch param.Position {
	case positionHeader:
		return header.Get(param.Key), nil
	case positionBody:
		result := param.expr.Get(body)
		if result == nil {
			return "", fmt.Errorf("fail to retrieve response param,key: %s,position: %s", param.Key, param.Position)
		}
		return result[0], nil
	}
	return "", fmt.Errorf("fail to retrieve response param,key: %s,position: %s", param.Key, param.Position)
}

func (i *IntrospectClient) Send(ctx http_service.IHttpContext) (*BaseResponse, error) {
	query := url.Values{}
	for k, v := range i.Query {
		query.Set(k, v)
	}
	endpoint := i.Endpoint
	if len(query) > 0 {
		endpoint = fmt.Sprintf("%s?%s", i.Endpoint, query.Encode())
	}
	var body string
	contentType := "application/x-www-form-urlencoded"
	switch i.ContentType {
	case FormData:
		b := url.Values{}
		for k, v := range i.Body {
			b.Set(k, v)
		}
		body = b.Encode()
	case Json:
		b, _ := json.Marshal(i.Body)
		body = string(b)
		contentType = "application/json"
	}

	req, err := http.NewRequest(i.Method, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range i.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &BaseResponse{
		Header: resp.Header,
		Status: resp.StatusCode,
		Body:   responseBody,
	}, nil
}
