package output

import (
	"github.com/eolinker/eosc"
)

const OutputSkill = "github.com/eolinker/apinto/http-entry.http-entry.IOutput"

type IEntryOutput interface {
	Output(entry eosc.IEntry) error
}

// CheckSkill 检查能力
func CheckSkill(skill string) bool {
	return skill == OutputSkill
}
