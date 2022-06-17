package basic

//Config basic配置内容
type Config struct {
	HideCredentials bool   `json:"hide_credentials" label:"是否隐藏证书"`
	User            []User `json:"user" label:"用户列表"`
}

//User 用户信息
type User struct {
	Username string            `json:"username" label:"用户名（username）" nullable:"false"`
	Password string            `json:"password" label:"密码（password）" nullable:"false"`
	Labels   map[string]string `json:"labels" label:"用户标签"`
	Expire   int64             `json:"expire" format:"date-time" label:"过期时间"`
}
