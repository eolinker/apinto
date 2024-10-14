package wenxin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	client = http.Client{}
)

func getToken(ak string, sk string) (*TokenResponse, error) {
	uri := "https://aip.baidubce.com/oauth/2.0/token"
	method := http.MethodPost
	query := url.Values{}
	query.Set("grant_type", "client_credentials")
	query.Set("client_id", ak)
	query.Set("client_secret", sk)
	req, err := http.NewRequest(method, uri, strings.NewReader(query.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response TokenResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type TokenResponse struct {
	//RefreshToken  string `json:"refresh_token"`
	ExpiresIn int `json:"expires_in"`
	//SessionKey    string `json:"session_key"`
	AccessToken string `json:"access_token"`
	//Scope         string `json:"scope"`
	//SessionSecret string `json:"session_secret"`
}
