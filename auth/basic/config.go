package basic

//Config basic配置内容
type Config struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	User   []User `json:"user"`
}

//User 用户信息
type User struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Labels   map[string]string `json:"labels"`
	Expire   int64             `json:"expire"`
}
