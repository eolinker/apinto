package apikey

//Config apiKey配置内容
type Config struct {
	HideCredentials bool   `json:"hide_credentials" label:"是否隐藏证书"`
	User            []User `json:"user" label:"用户列表"`
}

type apiKeyUsers struct {
	users []User
}

//User 用户信息
type User struct {
	Apikey string            `json:"apikey" label:"密钥（Apikey）" nullable:"false"`
	Labels map[string]string `json:"labels" label:"用户标签"`
	Expire int64             `json:"expire" format:"date-time" label:"过期时间"`
}
