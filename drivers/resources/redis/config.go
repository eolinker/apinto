package redis

type Config struct {
	Addrs    []string `json:"addrs" label:"redis 节点列表"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}
