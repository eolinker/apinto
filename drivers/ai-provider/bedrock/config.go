package bedrock

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	AccessKey          string `json:"aws_access_key_id"`
	SecretKey          string `json:"aws_secret_access_key"`
	Region             string `json:"aws_region"`
	ModelForValidation string `json:"model_for_validation"`
}

var (
	availableRegions = map[string]struct{}{
		"us-east-1":      {},
		"us-west-2":      {},
		"ap-southeast-1": {},
		"ap-northeast-1": {},
		"eu-central-1":   {},
		"eu-west-2":      {},
		"us-gov-west-1":  {},
		"ap-southeast-2": {},
	}
)

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.AccessKey == "" {
		return nil, fmt.Errorf("aws_access_key_id is required")
	}
	if conf.SecretKey == "" {
		return nil, fmt.Errorf("aws_secret_access_key is required")
	}
	//if conf.Region == "" {
	//	return nil, fmt.Errorf("aws_region is required")
	//}
	//if _, ok := availableRegions[conf.Region]; !ok {
	//	return nil, fmt.Errorf("aws_region is invalid")
	//}
	return conf, nil
}
