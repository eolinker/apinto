package filelog_transport

import "github.com/eolinker/eosc/log"

type Config struct {
	Dir    string
	File   string
	Expire int
	Period LogPeriod
	Level  log.Level
}
