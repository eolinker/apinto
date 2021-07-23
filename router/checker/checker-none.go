package checker

import "strings"

var (
	globalCheckerNone = &checkerNone{}
)
type checkerNone struct {

}

func (t *checkerNone) Key() string {
	return "$"
}

func (t *checkerNone) Value() string {
	return "$"
}

func newCheckerNone() *checkerNone {
	return globalCheckerNone
}

func (t *checkerNone) Check(v string, has bool) bool {
	if !has{
		return false
	}
	return strings.TrimSpace(v) == ""
}

func (t *checkerNone) CheckType() CheckType {
	return CheckTypeNone
}

