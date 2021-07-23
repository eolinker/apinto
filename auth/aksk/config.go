package aksk

type Config struct {
	Ak     string            `json:"ak"`
	Sk     string            `json:"sk"`
	Labels map[string]string `json:"labels"`
	Expire int64             `json:"expire"`
}
