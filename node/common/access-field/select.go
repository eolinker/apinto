package access_field

import "reflect"

var (
	defaultFields []string
)

func Default() []string {
	return defaultFields
}
func init() {
	defaultFields = initFieldS()
}
func initFieldS() []string {
	v := reflect.ValueOf(new(Fields)).Elem()
	t := v.Type()
	n := t.NumField()
	m := make([]string, 0, n)
	for i := 0; i < n; i++ {
		structField := t.Field(i)
		fn := structField.Tag.Get("field")
		m = append(m, fn)
	}
	return m
}
