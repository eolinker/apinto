package bedrock

import "fmt"

type Config struct {
	AccessKey string `json:"aws_access_key_id"`
	SecretKey string `json:"aws_secret_access_key"`
	Region    string `json:"aws_region"`
}

func checkConfig(conf *Config) error {
	if conf.AccessKey == "" {
		return fmt.Errorf("aws_access_key_id is required")
	}
	if conf.SecretKey == "" {
		return fmt.Errorf("aws_secret_access_key is required")
	}
	return nil
}
