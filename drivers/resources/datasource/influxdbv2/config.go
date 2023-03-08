package influxdbv2

type Config struct {
	Scopes []string `json:"scopes" label:"作用域"`
	Url    string   `json:"url"`
	Org    string   `json:"org"`
	Bucket string   `json:"bucket"`
	Token  string   `json:"token"`
}

func checkConfig(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, errorConfigType
	}
	return cfg, nil
}
