package professions

import (
	"github.com/eolinker/eosc/process-master/professions"
)

const (
	SpaceProfession = "profession"
)

type Professions struct {
	*professions.Professions
	fileName string
}

func (p *Professions) ResetHandler(data []byte) error {
	professions, err := readProfessionConfig(p.fileName)
	if err != nil {
		return err
	}
	p.Professions.Reset(professions)
	return nil
}

func NewProfessions(fileName string) *Professions {
	return &Professions{
		Professions: professions.NewProfessions(),
		fileName:    fileName,
	}
}
