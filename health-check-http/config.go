package health_check_http

import "time"

type Config struct {
	Protocol    string
	Method      string
	Url         string
	SuccessCode int
	Period      time.Duration
	Timeout     time.Duration
}
