package basic

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"liujian-test/grpc-test-demo/common/flag"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var names = []string{"basic", "basic-auth"}

type Auth struct {
}

func NewAuth() *Auth {
	return &Auth{}
}

func (a *Auth) Names() []string {
	return names
}

func (a *Auth) Auth(md map[string][]string) (string, error) {

	v, ok := md[":authority"]
	if !ok {
		return "", grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	authority := strings.Replace(v[0], "Basic ", "", 1)
	b, err := base64.StdEncoding.DecodeString(authority)
	if err != nil {
		return "", grpc.Errorf(codes.Unauthenticated, "token解析失败，格式错误："+err.Error())
	}

	infos := strings.Split(string(b), ":")
	username := infos[0]
	password := ""
	if len(infos) > 1 {
		password = infos[1]
	}
	if !checkPassword(username, password) {
		return "", errors.New("illegal password")
	}

	return username, nil
}

func checkPassword(username string, password string) bool {
	uri := fmt.Sprintf("%s/api/basic", flag.ConfigAddress)
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	type response struct {
		Data map[string]map[string]string `json:"data"`
	}
	data := new(response)
	err = json.Unmarshal(body, data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if v, ok := data.Data["basic"]; ok {
		if t, ok := v[username]; ok {
			if t == password {
				return true
			}
		}
	}

	return false
}
