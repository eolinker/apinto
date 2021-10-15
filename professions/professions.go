package professions

import (
	"github.com/eolinker/eosc/process-master/professions"
)

const (
	SpaceProfession = "profession"
)

type Professions struct {
	*professions.Professions
}

func NewProfessions(fileName string) (*Professions, error) {
	psConfig, err := readProfessionConfig(fileName)
	if err != nil {
		return nil, err
	}
	p := &Professions{
		Professions: professions.NewProfessions(),
	}
	p.Professions.Reset(psConfig)
	return p, nil
}
