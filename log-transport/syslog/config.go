package syslog

import "github.com/eolinker/eosc/log"

//Config syslog-Transporter所需配置
type Config struct {
	Network string
	RAddr   string
	Level   log.Level
}
