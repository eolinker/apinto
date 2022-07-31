package proxy_rewrite

var (
	regexpURIErrInfo = `[plugin proxy-rewrite config err] regex_uri's length must be 0 or 2. err regexURI: %s `
	regexpErrInfo    = `[plugin proxy-rewrite config err] compile regex_uri fail. err regexp: %s `
)
