package response_rewrite

import "fmt"

type Config struct {
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	BodyBase64 bool              `json:"body_base64"`
	Headers    map[string]string `json:"headers"`
	Match      *MatchConf        `json:"match"`
}

type MatchConf struct {
	Code []int `json:"code"`
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
