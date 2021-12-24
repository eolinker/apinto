package output

import "github.com/eolinker/eosc"

type IEntryOutput interface {
	Output(entry eosc.IEntry) error
}
