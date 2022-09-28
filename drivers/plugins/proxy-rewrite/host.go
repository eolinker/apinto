package proxy_rewrite

import (
	"github.com/eolinker/eosc/eocontext"
)

type upstreamHostRewrite string

func (u upstreamHostRewrite) PassHost() (eocontext.PassHostMod, string) {
	return eocontext.ReWriteHost, string(u)
}
