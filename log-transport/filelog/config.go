package filelog

import "github.com/eolinker/eosc/log"

//Config filelog-Transporter所需配置
type Config struct {
	Dir    string
	File   string
	Expire int
	Period LogPeriod
	Level  log.Level
}
