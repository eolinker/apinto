package log

import "github.com/eolinker/eosc/log"

//CheckSkill 检查能力
func CheckSkill(skill string) bool {
	return false
}

//ITransporter transporter接口声明
type ITransporter interface {
	log.EntryTransporter
}

type TransporterReset interface {
	log.EntryTransporter
	Reset(config interface{}, formatter log.Formatter) error
}
