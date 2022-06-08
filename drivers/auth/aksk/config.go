package aksk

type Config struct {
	Name            string       `json:"name"`
	Driver          string       `json:"driver"`
	HideCredentials bool         `json:"hide_credentials"`
	Users           []AKSKConfig `json:"user"`
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
