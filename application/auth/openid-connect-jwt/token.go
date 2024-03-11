package openid_connect_jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type tokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

func extractTokenHeader(token string) (*tokenHeader, error) {
	ts := strings.Split(token, ".")
	if len(ts) != 3 {
		return nil, ErrInvalidToken
	}
	headerData, err := base64.RawStdEncoding.DecodeString(ts[0])
	if err != nil {
		return nil, err
	}
	var th tokenHeader
	err = json.Unmarshal(headerData, &th)
	if err != nil {
		return nil, err
	}
	return &th, nil
}
