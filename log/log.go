package log

import "github.com/eolinker/eosc/log"

//CheckSkill 检查能力
func CheckSkill(skill string) bool {
	return false
}

type TransporterReset interface {
	log.EntryTransporter
	Reset(config interface{}, formatter log.Formatter) error
}
