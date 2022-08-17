package application

import (
	"github.com/eolinker/eosc/utils/config"
)

var (
	appSkill string
)

func init() {
	var t IApp
	appSkill = config.TypeNameOf(&t)
}

type IApp interface {
}

func CheckSkill(skill string) bool {
	return skill == appSkill
}
