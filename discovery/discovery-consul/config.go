package discovery_consul

type Config struct {
	Name   string            `json:"name"`
	Driver string            `json:"driver"`
	Labels map[string]string `json:"labels"`
	Config AccessConfig      `json:"config"`
}

type AccessConfig struct {
	Address []string          `json:"address"`
	Params  map[string]string `json:"params"`
}
