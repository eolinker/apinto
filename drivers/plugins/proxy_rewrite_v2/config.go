package proxy_rewrite_v2

type Config struct {
	PathType    string            `json:"path_type" label:"path重写类型" enum:"none,static,prefix,regex"`
	StaticPath  string            `json:"static_path" label:"静态path" switch:"path_type==='static'"`
	PrefixPath  []*SPrefixPath    `json:"prefix_path" label:"path前缀替换" switch:"path_type==='prefix'"`
	RegexPath   []*SRegexPath     `json:"regex_path" label:"path正则替换" switch:"path_type==='regex'"`
	NotMatchErr bool              `json:"not_match_err" label:"path替换失败是否报错"`
	HostRewrite bool              `json:"host_rewrite" label:"是否重写host"`
	Host        string            `json:"host" label:"Host" switch:"host_rewrite===true"`
	Headers     map[string]string `json:"headers" label:"请求头部"`
}

type SPrefixPath struct {
	PrefixPathMatch   string `json:"prefix_path_match" label:"path前缀匹配字符串"`
	PrefixPathReplace string `json:"prefix_path_replace" label:"path前缀替换字符串"`
}

type SRegexPath struct {
	RegexPathMatch   string `json:"regex_path_match" label:"path正则匹配表达式"`
	RegexPathReplace string `json:"regex_path_replace" label:"path正则替换表达式"`
}

func (c *Config) doCheck() error {
	switch c.PathType {
	case typeStatic, typePrefix, typeRegex:
	default:
		c.PathType = typeNone
	}

	//if c.HostRewrite && len(c.Host) == 0 {
	//	return fmt.Errorf(hostErrInfo)
	//}

	return nil
}
