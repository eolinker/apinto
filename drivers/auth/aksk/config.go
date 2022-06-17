package aksk

type Config struct {
	HideCredentials bool         `json:"hide_credentials" label:"是否隐藏证书"`
	Users           []AKSKConfig `json:"user" label:"用户列表"`
}

type akskUsers struct {
	users []AKSKConfig
}

type AKSKConfig struct {
	AK     string            `json:"ak" label:"Access Key" nullable:"false"`
	SK     string            `json:"sk" label:"Secret Access Key" nullable:"false"`
	Labels map[string]string `json:"labels" label:"用户标签"`
	Expire int64             `json:"expire" format:"date-time" label:"过期时间"`
}
