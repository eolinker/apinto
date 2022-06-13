package aksk

type Config struct {
	HideCredentials bool         `json:"hide_credentials" label:"是否隐藏证书"`
	Users           []AKSKConfig `json:"user" label:"用户列表"`
}

type akskUsers struct {
	users []AKSKConfig
}

type AKSKConfig struct {
	AK     string            `json:"ak"`
	SK     string            `json:"sk"`
	Labels map[string]string `json:"labels"`
	Expire int64             `json:"expire" format:"date-time"`
}
