package checker

import "strings"

type checkerNone struct {

}

func (t *checkerNone) Key() string {
	return "$"
}

func (t *checkerNone) Value() string {
	return "$"
}

func newCheckerNone() *checkerNone {
	return &checkerNone{}
}

func (t *checkerNone) Check(v string, has bool) bool {
	if !has{
		return false
	}
	return strings.TrimSpace(v) == ""
}

func (t *checkerNone) CheckType() CheckType {
	return CheckTypeNull
}

