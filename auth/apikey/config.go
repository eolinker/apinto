package apikey

//Config basic配置内容
type Config struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
	User   []User `json:"user"`
}

//User 用户信息
type User struct {
	Apikey string            `json:"apikey"`
	Label  map[string]string `json:"label"`
	Expire int64             `json:"expire"`
}
