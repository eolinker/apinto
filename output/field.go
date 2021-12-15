package output

var (
	HttpPrefix        = "http_"
	QueryPrefix       = "query_"
	ProxyQueryPrefix  = "proxy_query_"
	ProxyHeaderPrefix = "proxy_header_"
	CookiePrefix      = "cookie_"
	Prefixes          = []string{
		HttpPrefix,
		QueryPrefix,
		ProxyQueryPrefix,
		ProxyHeaderPrefix,
	}

	//KeyURI 请求中的当前URI(不带请求参数，参数位于$args)，可以不同于浏览器传递的$request_uri的值，它可以通过内部重定向，或者使用index指令进行修改，$uri不包含主机名，如"/foo/bar.html"
	KeyURI = "uri"

	//KeyContentLength "Content-Length" 请求头字段
	KeyContentLength = "content_length"

	//KeyContentType "Content-Type" 请求头字段
	KeyContentType = "content_type"

	//KeyMsec "msec"
	KeyMsec = "msec"

	//KeyApintoVersion apinto版本
	KeyApintoVersion = "apinto_version"

	//KeyRemoteAddr 客户端地址
	KeyRemoteAddr = "remote_addr"

	//KeyRemotePort 客户端端口
	KeyRemotePort = "remote_port"

	//KeyRequest 客户端的请求地址
	KeyRequest = "request"

	//KeyRequestBody 客户端的请求主体
	KeyRequestBody = "request_body"

	//KeyRequestLength 请求的长度 (包括请求的地址，http请求头和请求主体)
	KeyRequestLength = "request_length"

	//KeyRequestMethod HTTP请求方法，通常为 "GET" 或 "POST"
	KeyRequestMethod = "request_method"

	//KeyRequestTime 处理客户端请求使用的时间,单位为秒，精度毫秒； 从读入客户端的第一个字节开始，直到把最后一个字符发送给客户端后进行日志写入为止。
	KeyRequestTime = "request_time"

	//KeyRequestUri 这个变量等于包含一些客户端请求参数的原始URI
	KeyRequestUri = "request_uri"

	//KeyScheme 请求使用的Web协议，"http" 或 "https"
	KeyScheme = "scheme"

	//KeyStatus HTTP响应状态
	KeyStatus = "status"

	//KeyTimeISO8601 服务器时间的ISO 8601格式
	KeyTimeISO8601 = "time_iso8601"

	//KeyTimeLocal 服务器时间（“2006-01-02 15:04:05”格式）
	KeyTimeLocal = "time_local"

	//KeyHeader 所有请求头字段，按照原格式输出
	KeyHeader = "header"

	//KeyHost 请求地址，即浏览器中你输入的地址（IP或域名，不包括端口）
	KeyHost = "host"

	//KeyProxyHeader 转发请求的请求头，该值按照原格式输出所有头部信息
	KeyProxyHeader = "proxy_header"

	//KeyProxyQuery 转发请求的所有query参数，该值进行url encode编码
	KeyProxyQuery = "proxy_query"

	//KeyProxyUri 转发请求的uri
	KeyProxyUri = "proxy_uri"

	//KeyProxyScheme 转发请求的协议
	KeyProxyScheme = "proxy_scheme"

	//KeyProxyBody 转发请求的请求体
	KeyProxyBody = "proxy_body"

	//KeyProxyHost 上游服务的host地址（IP或域名，不包括端口）
	KeyProxyHost = "proxy_host"

	//KeyProxyPort 上游服务的端口
	KeyProxyPort = "proxy_port"
)
