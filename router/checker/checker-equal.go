package checker

import "fmt"

type checkerEqual struct {
	value string
}

func (e *checkerEqual) Key() string {
	return fmt.Sprintf("=%s",e.value)
}

func (e *checkerEqual) Value() string {
	return e.value
}

func newCheckerEqual(value string) *checkerEqual {
	return &checkerEqual{value: value}
}

func (e *checkerEqual) Check(v string,has bool) bool{
	if !has{
		return false
	}
	return v == e.value
}

func (e *checkerEqual) CheckType() CheckType {
	return CheckTypeEqual
}
