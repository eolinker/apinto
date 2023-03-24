package discovery_old

// IHealthCheckerFactory 健康检查工厂类接口
type IHealthCheckerFactory interface {
	IHealthChecker
	Agent() (IHealthChecker, error)
	Reset(conf interface{}) error
}

// IHealthChecker 健康检查接口
type IHealthChecker interface {
	AddToCheck(node BaseNode) error
	Stop() error
}
