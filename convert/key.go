package convert

//type IKeySelector interface {
//	Next() (IKeyResource, bool)
//}

type IKeyResource interface {
	ID() string
	Health() bool
	Priority() int
	// Up 上线
	Up()
	// Down 下线
	Down()

	IsBreaker() bool
	// Breaker 熔断
	Breaker()
	ConverterDriver() IConverterDriver
}
