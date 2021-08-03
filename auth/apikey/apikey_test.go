package apikey

import (
	"bytes"
	"encoding/json"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)
import "errors"

var (
	users = []User{
		{
			Apikey: "asdqer",
			Label:  make(map[string]string),
		},
		{
			Apikey: "eolinker",
			Label:  make(map[string]string),
			Expire: 0,
		},
		{
			Apikey: "goku",
			Label:  make(map[string]string),
			Expire: 1627013522,
		},
	}
	cfg = &Config{
		Name:   "apikey_test",
		Driver: "apikey",
		User:   users,
	}
)

func TestHeaderAuthorization(t *testing.T) {
	worker, err := getWorker("", "AuthorizationType")
	if err != nil {
		t.Error(err)
		return
	}
	headers := map[string]string{
		"authorization-type": "Apikey",
		"authorization":      "eolinker",
	}
	req, err := buildRequest(headers, nil, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = worker.Auth(http_context.NewContext(req, &writer{}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("auth success")
	return
}
func TestQueryAuthorization(t *testing.T) {
	worker, err := getWorker("", "AuthorizationType")
	if err != nil {
		t.Error(err)
		return
	}
	headers := map[string]string{
		"authorization-type": "Apikey",
	}
	query := map[string]string{
		"Apikey": "eolinker",
	}
	req, err := buildRequest(headers, query, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = worker.Auth(http_context.NewContext(req, &writer{}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("auth success")
	return
}
func TestBodyAuthorization(t *testing.T) {
	var jsonBody = &struct {
		Apikey string
	}{
		Apikey: "eolinker",
	}

	body, err := json.Marshal(jsonBody)
	if err != nil {
		t.Error(err)
		return
	}
	worker, err := getWorker("", "AuthorizationType")
	if err != nil {
		t.Error(err)
		return
	}
	headers := map[string]string{
		"authorization-type": "Apikey",
		"Content-Type":       "application/json",
	}
	req, err := http.NewRequest(http.MethodPost, "localhost:8081", bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	err = worker.Auth(http_context.NewContext(req, &writer{}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("auth success")
	return
}

func TestMultipartFormAuthorization(t *testing.T) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	err := w.WriteField("Apikey", "eolinker")
	if err != nil {
		t.Error(err)
		return
	}
	w.Close()
	worker, err := getWorker("", "AuthorizationType")
	if err != nil {
		t.Error(err)
		return
	}
	headers := map[string]string{
		"authorization-type": "Apikey",
	}
	req, err := http.NewRequest(http.MethodPost, "localhost:8081", buf)
	if err != nil {
		t.Error(err)
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	err = worker.Auth(http_context.NewContext(req, &writer{}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("auth success")
	return
}
func TestFormAuthorization(t *testing.T) {
	var formBody = url.Values{
		"Apikey": []string{"eolinker"},
	}
	worker, err := getWorker("", "AuthorizationType")
	if err != nil {
		t.Error(err)
		return
	}
	headers := map[string]string{
		"authorization-type": "Apikey",
		"Content-Type":       "application/x-www-form-urlencoded",
	}
	req, err := buildRequest(headers, nil, formBody.Encode())
	if err != nil {
		t.Error(err)
		return
	}
	err = worker.Auth(http_context.NewContext(req, &writer{}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("auth success")
	return
}

func getWorker(id string, name string) (auth.IAuth, error) {
	f := NewFactory()
	driver, err := f.Create("auth", "apikey", "", "apikey驱动", nil)
	if err != nil {
		return nil, err
	}
	worker, err := driver.Create(id, name, cfg, nil)
	if err != nil {
		return nil, err
	}
	a, ok := worker.(auth.IAuth)
	if !ok {
		return nil, errors.New("invalid struct type")
	}
	return a, nil
}

func buildRequest(headers map[string]string, query map[string]string, body string) (*http.Request, error) {
	method := http.MethodPost
	if len(query) > 0 {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, "localhost:8081", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if len(query) > 0 {
		params := make(url.Values)
		for key, value := range query {
			params.Add(key, value)
		}
		req.URL.RawQuery = params.Encode()
	}
	return req, err
}

type writer struct {
}

func (w writer) Header() http.Header {
	panic("implement me")
}

func (w writer) Write(bytes []byte) (int, error) {
	panic("implement me")
}

func (w writer) WriteHeader(statusCode int) {
	panic("implement me")
}
