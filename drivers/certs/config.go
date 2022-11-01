package certs

type Config struct {
	Name string `json:"name" label:"证书名"`
	Key  string `json:"key" label:"key"`
	Pem  string `json:"pem" label:"value"`
}
