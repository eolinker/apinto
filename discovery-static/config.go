package discovery_static

type Config struct {
	Name     string            `json:"name"`
	Driver   string            `json:"driver"`
	Labels   map[string]string `json:"labels"`
	Health   *HealthConfig     `json:"health"`
	HealthOn bool              `json:"health_on"`
}

type AccessConfig struct {
	Address []string          `json:"address"`
	Params  map[string]string `json:"params"`
}

type HealthConfig struct {
	Protocol    string `json:"protocol"`
	Method      string `json:"method"`
	Url         string `json:"url"`
	SuccessCode int    `json:"success_code"`
	Period      int    `json:"period"`
	Timeout     int    `json:"timeout"`
}
