package checker

import "fmt"

type checkerNotEqual struct {
	value string
}

func (e *checkerNotEqual) Key() string {
	return fmt.Sprintf("!=%s",e.value)
}

func (e *checkerNotEqual) Value() string {
	return e.value
}

func newCheckerNotEqual(value string) *checkerNotEqual {
	return &checkerNotEqual{value: value}
}

func (e *checkerNotEqual) Check(v string,has bool) bool{
	if !has{
		return false
	}
	return v != e.value
}

func (e *checkerNotEqual) CheckType() CheckType {
	return CheckTypeNotEqual
}
