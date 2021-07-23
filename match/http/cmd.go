package http

import "github.com/eolinker/goku-eosc/match"

const(
	cmdLocation ="LOCALTION:"
	cmdHeader = "HEADER___:"
	cmdCookie = "COOKIE___:"
	cmdQuery = "QUERY____:"
	cmdHost = "HOST_____:"
)

var (
	cmdFunc = map[string]func(v string)match.IReader{
		cmdLocation: func(v string) IReader {
			return LocationReader(v)
		},
		cmdHeader: func(v string) IReader {
			return HeaderReader(v)
		},
		cmdCookie: func(v string) IReader {
			return CookieReader(v)
		},
		cmdQuery: func(v string) IReader {
			return QueryReader(v)
		},
		cmdHost: func(v string) IReader {
			return HostReader(v)
		},
	}
)

