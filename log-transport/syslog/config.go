package syslog

import "github.com/eolinker/eosc/log"

type Config struct {
	Network string
	RAddr   string
	Level   log.Level
}
