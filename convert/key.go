package convert

type IKeyPool interface {
	Provider() string
	Model() string
	ModelConfig() map[string]interface{}
	Priority() int
	Health() bool
	Selector() IKeySelector
	Down()
}

type IKeySelector interface {
	Next() (IKeyResource, bool)
}

type IKeyResource interface {
	Health() bool
	// Up 上线
	Up()
	// Down 下线
	Down()
	// Breaker 熔断
	Breaker()
	ConverterDriver() IConverterDriver
}
