package healthcheckhttp

import "time"

//Config healthCheck所需配置
type Config struct {
	Protocol    string
	Method      string
	URL         string
	SuccessCode int
	Period      time.Duration
	Timeout     time.Duration
}
