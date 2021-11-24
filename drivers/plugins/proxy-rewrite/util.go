package proxy_rewrite

var (
	schemeErrInfo    = `[plugin proxy-rewrite config err] scheme must be in the set ["http","https"]. err scheme: %s `
	regexpURIErrInfo = `[plugin proxy-rewrite config err] regex_uri's length must be 0 or 2. err regexURI: %s `
	regexpErrInfo    = `[plugin proxy-rewrite config err] compile regex_uri fail. err regexp: %s `
	uriErrInfo       = `[plugin proxy-rewrite config err] uri or regexUri at least exist one item`
)
