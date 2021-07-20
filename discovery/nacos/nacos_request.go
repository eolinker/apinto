package nacos

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// nacos 实例结构
type Instance struct {
	Hosts []struct {
		Valid      bool    `json:"valid"`
		Marked     bool    `json:"marked"`
		InstanceId string  `json:"instanceId"`
		Port       int     `json:"port"`
		Ip         string  `json:"ip"`
		Weight     float64 `json:"weight"`
	}
}

func MapToJson(param map[string]interface{}) string {
	dataType, _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}

func SendRequest(method string, request string, query map[string]string, body map[string]string) (*http.Response, error) {
	// 构造url参数字符串
	paramsUrl := url.Values{}
	for key, value := range query {
		paramsUrl.Add(key, value)
	}
	// 更新url
	paramsUrlString := paramsUrl.Encode()
	if paramsUrlString != "" {
		request = request + "?" + paramsUrlString
	}
	var bodyReader io.Reader
	// 构造urlencoded请求体
	if method == http.MethodPost || method == http.MethodPut {
		bodyUrl := url.Values{}
		for key, value := range body {
			bodyUrl.Add(key, value)
		}
		bodyUrlString := bodyUrl.Encode()
		bodyReader = strings.NewReader(bodyUrlString)
	}
	req, err := http.NewRequest(method, request, bodyReader)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// get nacos query parameters
func (n *nacos) getParams(serviceName string) map[string]string {
	query := n.params
	query["serviceName"] = serviceName
	if _, ok := query["healthyOnly"]; !ok {
		query["healthyOnly"] = "true"
	}
	return query
}
