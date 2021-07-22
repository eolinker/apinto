package utils

import (
	"encoding/json"
	"fmt"

	"github.com/robertkrimen/otto"
)

//JSObjectToJSON 将js对象转为json
func JSObjectToJSON(s string) ([]byte, error) {
	vm := otto.New()
	v, err := vm.Run(fmt.Sprintf(`
		cs = %s
		JSON.stringify(cs)
`, s))
	if err != nil {
		return nil, err
	}
	return []byte(v.String()), nil
}
func JsonUnmarshal(s, v interface{}) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
