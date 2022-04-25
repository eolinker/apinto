package professions

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/process-master/professions"
)

const (
	SpaceProfession = "profession-admin"
)

type Professions struct {
	*professions.Professions
}

func NewProfessions() (*Professions, error) {

	p := &Professions{
		Professions: professions.NewProfessions(),
	}
	p.Professions.Reset(ApintoProfession())
	return p, nil
}

func (p *Professions) Reset([]*eosc.ProfessionConfig) {
	return
}
