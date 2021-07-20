package discover_eureka


type Config struct {
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Labels   map[string]string
	Config   AccessConfig `json:"config"`
}

type AccessConfig struct {
	Address []string
	Params  map[string]string
}
