package openid_connect_jwt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type issuerHandler struct {
	prefix string
}

func newIssuerHandler(prefix string) *issuerHandler {
	return &issuerHandler{prefix: strings.TrimPrefix(prefix, "/")}
}

func (h *issuerHandler) PrefixPath() string {
	return h.prefix
}

func (h *issuerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, fmt.Sprintf("/apinto/%s", h.prefix))
	switch request.Method {
	case http.MethodGet:
		if path == "" {
			h.list(writer, request)
			return
		}
		paths := strings.SplitN(path, "/", 2)
		if len(paths) == 2 {
			h.info(writer, request, paths[1])
			return
		}
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("not found"))
		return

	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("method not allowed"))
		return
	}
}

func (h *issuerHandler) info(writer http.ResponseWriter, request *http.Request, id string) {
	var body []byte
	info, has := manager.Issuers.Get(id)
	if !has {
		body = []byte("{\"message\":\"not found\"}")
	} else {
		body, _ = json.Marshal(info)
	}
	writer.Write(body)

}

func (h *issuerHandler) list(writer http.ResponseWriter, request *http.Request) {
	var data = map[string]interface{}{
		"data": manager.Issuers.List(),
	}
	body, _ := json.Marshal(data)
	writer.Write(body)
}

type jwkHandler struct {
	prefix string
}

func newJwkHandler(prefix string) *jwkHandler {
	return &jwkHandler{prefix: strings.TrimPrefix(prefix, "/")}
}

func (j *jwkHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, fmt.Sprintf("/apinto/%s", j.prefix))
	switch request.Method {
	case http.MethodGet:
		if path == "" {
			j.list(writer, request)
			return
		}
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("not found"))
		return

	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("method not allowed"))
		return
	}
}

func (j *jwkHandler) list(writer http.ResponseWriter, request *http.Request) {

	all := manager.Issuers.List()
	keys := make([]JWK, 0, len(all))
	for _, issuer := range all {
		keys = append(keys, issuer.Keys...)
	}
	var data = map[string]interface{}{
		"keys": keys,
	}
	body, _ := json.Marshal(data)
	writer.Write(body)
}
