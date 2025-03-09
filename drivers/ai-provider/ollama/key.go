package ollama

type key struct {
	id            string
	convertDriver ai_convert.IConverterDriver
}

func newKey(id string, convertDriver ai_convert.IConverterDriver) ai_convert.IKeyResource {
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

func (k *key) ConverterDriver() ai_convert.IConverterDriver {
	return k.convertDriver
}
