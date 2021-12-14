package output

import (
	"github.com/eolinker/eosc/formatter"
)

type IOutput interface {
	Output(entry formatter.IEntry) error
}
