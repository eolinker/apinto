package checker

type checkerAll struct {

}

func (t *checkerAll) Key() string {
	return "*"
}

func (t *checkerAll) Value() string {
	return "*"
}

func newCheckerAll() *checkerAll {
	return &checkerAll{}
}

func (t *checkerAll) Check(v string, has bool) bool {
	return true
}

func (t *checkerAll) CheckType() CheckType {
	return CheckTypeAll
}

