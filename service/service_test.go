package service

import (
	"fmt"
	"reflect"
	"testing"
)

func TestService(t *testing.T) {
	fmt.Println(TypeNameOf((*IService)(nil)))
}

func TypeNameOf(v interface{}) string {
	return TypeName(reflect.TypeOf(v))
}

func TypeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return TypeName(t.Elem())
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.String())
}
