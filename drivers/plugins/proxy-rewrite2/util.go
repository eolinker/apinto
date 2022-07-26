package proxy_rewrite2

var (
	schemeErrInfo    = `[plugin proxy-rewrite2 config err] scheme must be in the set ["http","https"]. err scheme: %s `
	regexpURIErrInfo = `[plugin proxy-rewrite2 config err] regex_uri's length must be 0 or 2. err regexURI: %s `
	regexpErrInfo    = `[plugin proxy-rewrite2 config err] compile regexp fail. err regexp: %s `
	uriErrInfo       = `[plugin proxy-rewrite2 config err] uri or regexUri at least exist one item`

	hostErrInfo = `[plugin proxy-rewrite2 config err] host can't be null. `
)
