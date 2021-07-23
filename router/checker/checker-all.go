package checker

var globalCheckerAll = &checkerAll{}
type checkerAll struct {

}

func (t *checkerAll) Key() string {
	return "*"
}

func (t *checkerAll) Value() string {
	return "*"
}

func (t *checkerAll) Check(v string, has bool) bool {
	return true
}

func (t *checkerAll) CheckType() CheckType {
	return CheckTypeAll
}

func newCheckerAll() *checkerAll {
	return globalCheckerAll
}
