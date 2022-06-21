package response_rewrite

import "fmt"

type Config struct {
	StatusCode int               `json:"status_code" label:"响应状态码" minimum:"100" description:"最小值：100"`
	Body       string            `json:"body" label:"响应内容"`
	BodyBase64 bool              `json:"body_base64" label:"是否base64加密"`
	Headers    map[string]string `json:"headers" label:"响应头部"`
	Match      *MatchConf        `json:"match" label:"匹配状态码列表"`
}

type MatchConf struct {
	Code []int `json:"code" label:"状态码" minimum:"100" description:"最小值：100"`
}

var (
	statusCodeErrInfo = "[plugin response-rewrite config err] status_code must be in the range of [200,598]. err status_code: %d. "
	matchErrInfo      = "[plugin response-rewrite config err] match must be not null. "
	matchCodeErrInfo  = "[plugin response-rewrite config err] match's code is illegal. "
)

func (c *Config) doCheck() error {
	//status_code不填则为整型的默认值0，若响应状态码为默认值，且范围不在[200,598]  返回报错
	if c.StatusCode != 0 && (c.StatusCode < 200 || c.StatusCode > 598) {
		return fmt.Errorf(statusCodeErrInfo, c.StatusCode)
	}

	//match必填
	if c.Match == nil {
		return fmt.Errorf(matchErrInfo)
	}

	//match内的code数组不能为空，且有效范围为[200,598]
	if len(c.Match.Code) == 0 {
		return fmt.Errorf(matchCodeErrInfo)
	}
	for _, v := range c.Match.Code {
		if v < 200 || v > 598 {
			return fmt.Errorf(statusCodeErrInfo, c.StatusCode)
		}
	}

	return nil
}
