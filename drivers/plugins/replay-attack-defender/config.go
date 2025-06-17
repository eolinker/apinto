package replay_attack_defender

import "github.com/eolinker/eosc"

type Config struct {
	NonceHeader       string         `json:"nonce_header" label:"API调用者生成的唯一UUID存放的请求头" default:"X-Ca-Nonce"`
	TimestampHeader   string         `json:"timestamp_header" label:"10位时间戳存放的请求头" default:"X-Ca-Timestamp"`
	SignHeader        string         `json:"sign_header" label:"防重放签名存放的请求头" default:"X-Ca-Signature"`
	ReplayAttackToken string         `json:"replay_attack_token" label:"重放攻击防御令牌" default:"apinto"`
	TTL               int            `json:"ttl" label:"过期时间，单位：s" default:"600"`
	Cache             eosc.RequireId `json:"cache" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" label:"缓存位置"`
}
