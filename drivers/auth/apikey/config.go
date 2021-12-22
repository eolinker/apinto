package apikey

//Config apiKey配置内容
type Config struct {
	Driver          string `json:"driver"`
	Name            string `json:"name"`
	HideCredentials bool   `json:"hide_credentials"`
	User            []User `json:"user"`
}

type apiKeyUsers struct {
	users []User
}

//User 用户信息
type User struct {
	Apikey string            `json:"apikey"`
	Label  map[string]string `json:"label"`
	Expire int64             `json:"expire"`
}
