package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"liujian-test/grpc-test-demo/common/flag"
	"net/http"
)

var names = []string{"jwt", "jwt-auth"}

type Auth struct {
}

func NewAuth() *Auth {
	return &Auth{}
}

func (a *Auth) Names() []string {
	return names
}

func (a *Auth) Auth(md map[string][]string) (string, error) {
	return check(md)
}

func check(md map[string][]string) (string, error) {
	uri := fmt.Sprintf("%s/api/jwt", flag.ConfigAddress)
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type response struct {
		Data map[string]map[string]*Conf `json:"data"`
	}
	data := new(response)
	err = json.Unmarshal(body, data)
	if err != nil {
		return "", err
	}
	if v, ok := data.Data["jwt"]; ok {
		user, err := DoJWTAuthentication(v, md)
		if err != nil {
			return "", err
		}
		return user, nil
	}

	return "", errors.New("no jwt authorization")
}
