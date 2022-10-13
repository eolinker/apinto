package redis

type Config struct {
	Enable   bool     `json:"enable"`
	Addrs    []string `json:"addrs" label:"redis 节点列表"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}
