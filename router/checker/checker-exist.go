package checker

var (
	globalCheckerExist =  &checkerExist{}
	globalCheckerNotExist = &checkerNotExits{}
)
type checkerExist struct {

}

func (t *checkerExist) Key() string {
	return "**"
}

func (t *checkerExist) Value() string {
	return "**"
}

func newCheckerExist() *checkerExist {

	 return globalCheckerExist
}

func (t *checkerExist) Check(v string, has bool) bool {
	return has && len(v)>0
}

func (t *checkerExist) CheckType() CheckType {
	return CheckTypeExist
}

type checkerNotExits struct {

}

func (c *checkerNotExits) Key() string {
	return "!"
}

func (c *checkerNotExits) Value() string {
	return "!"
}

func (c *checkerNotExits) Check(v string, has bool) bool {
	return !has
}

func (c *checkerNotExits) CheckType() CheckType {
	return CheckTypeNotExist
}

func newCheckerNotExits() *checkerNotExits {
	return globalCheckerNotExist
}