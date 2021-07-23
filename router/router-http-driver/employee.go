package router

func NewEmployee(id string, name string, driver string, config string) *Employee {
	return &Employee{
		id:     id,
		name:   name,
		driver: driver,
		config: config,
	}
}

type Employee struct {
	id     string
	name   string
	driver string
	config string
}

func (e *Employee) Technique() []string {
	return []string{}
}

func (e *Employee) ID() string {
	return e.id
}

func (e *Employee) Name() string {
	return e.name
}

func (e *Employee) Driver() string {
	return e.driver
}

func (e *Employee) Config() string {
	return e.config
}
