package nacos

import (
	"encoding/json"
	"net/http"
	"net/url"
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

func SendRequest(uri string, serviceName string) (*http.Response, error) {
	// 构造url参数字符串
	paramsUrl := url.Values{}
	paramsUrl.Set("serviceName", serviceName)
	paramsUrl.Set("healthyOnly", "true")

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = paramsUrl.Encode()
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
