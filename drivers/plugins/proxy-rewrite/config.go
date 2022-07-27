package proxy_rewrite

import "fmt"

type Config struct {
	Scheme   string            `json:"scheme" label:"协议(已废弃)"`
	URI      string            `json:"uri" label:"路径"`
	RegexURI []string          `json:"regex_uri" label:"正则替换路径（regex_uri）" description:"该数组需要配置两个正则，第一个是匹配正则，第二个是替换正则。"`
	Host     string            `json:"host" label:"Host"`
	Headers  map[string]string `json:"headers" label:"请求头部" description:"可对转发请求的头部进行新增，修改，删除。配置的kv对，不存在则新增，已存在则进行覆盖重写，但需要注意特殊头部字段只能在后面添加新值而不能覆盖。value为空字符串表示删除。"`
}

func (c *Config) doCheck() error {

	lenRegURI := len(c.RegexURI)

	//RegexURI切片要么为空，要么只有两个值,第一个值为正则匹配值，第二个是用于替换的正则字符串
	if lenRegURI > 0 && lenRegURI != 2 {
		return fmt.Errorf(regexpURIErrInfo, c.RegexURI)
	}

	return nil
}
