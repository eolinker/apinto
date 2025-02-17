package ollama

import (
	"github.com/eolinker/apinto/convert"
)

type key struct {
	id            string
	convertDriver convert.IConverterDriver
}

func newKey(id string, convertDriver convert.IConverterDriver) convert.IKeyResource {
	return &key{
		id:            id,
		convertDriver: convertDriver,
	}
}

func (k *key) ID() string {
	return k.id
}

func (k *key) Priority() int {
	return 0
}

func (k *key) IsBreaker() bool {
	return false
}

func (k *key) Health() bool {
	return true
}

func (k *key) Up() {
	return
}

func (k *key) Down() {
	return
}

func (k *key) Breaker() {
	return
}

func (k *key) ConverterDriver() convert.IConverterDriver {
	return k.convertDriver
}
