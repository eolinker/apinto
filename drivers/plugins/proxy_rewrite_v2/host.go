package proxy_rewrite_v2

import (
	"github.com/eolinker/eosc/eocontext"
)

type upstreamHostRewrite string

func (u upstreamHostRewrite) PassHost() (eocontext.PassHostMod, string) {
	return eocontext.ReWriteHost, string(u)
}
